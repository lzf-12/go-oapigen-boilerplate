package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"oapi-to-rest/pkg/db"
	"oapi-to-rest/pkg/errlib"
	"oapi-to-rest/pkg/jwt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthImpl struct {
	Db  *db.SQLite
	Jwt *jwt.TokenManager
}

var _ StrictServerInterface = (*AuthImpl)(nil)

func (a *AuthImpl) PostRegister(ctx context.Context, request PostRegisterRequestObject) (PostRegisterResponseObject, error) {

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(request.Body.Password), bcrypt.DefaultCost)
	if err != nil {
		return PostRegister500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeInternalServer)
	}

	tx, err := a.Db.DB.BeginTx(ctx, nil)
	if err != nil {
		return PostRegister500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeDBTransaction)
	}

	var userID int64
	result, err := tx.ExecContext(ctx, `
		INSERT INTO users (email, first_name, last_name)
		VALUES (?, ?, ?)`,
		request.Body.Email, request.Body.FirstName, request.Body.LastName,
	)
	if err != nil {
		tx.Rollback()
		return PostRegister500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeDBTransaction)
	}

	userID, err = result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return PostRegister500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeInternalServer)
	}

	// insert into auth_credentials table
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO auth_credentials (user_id, provider, password_hash)
		VALUES (?, 'local', ?)`,
		userID, string(hashPassword),
	); err != nil {
		tx.Rollback()
		return PostRegister500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeDBTransaction)
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return PostRegister500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeDBTransaction)
	}

	resp := RegisterResponse{
		Data: struct {
			Email *string "json:\"email,omitempty\""
		}{Email: &request.Body.Email},
		Message:    "success register new user",
		StatusCode: http.StatusCreated,
	}

	return PostRegister201JSONResponse(resp), nil
}

func (a *AuthImpl) PostLogin(ctx context.Context, request PostLoginRequestObject) (PostLoginResponseObject, error) {

	// check email exist
	row := a.Db.DB.QueryRowContext(ctx, `
	SELECT u.id, u.email, ac.password_hash FROM users u
	LEFT join auth_credentials ac 
	WHERE u.email = $1 order by created_at desc limit 1;
`, request.Body.Email)

	var userID int
	var email string
	var hashedPassword string
	if err := row.Scan(&userID, &email, &hashedPassword); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// email not found
			return PostLogin401JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeInvalidEmailOrPassword)
		}
		return PostLogin500JSONResponse{}, err
	}

	// compare hash with bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(request.Body.Password)); err != nil {

		if err == bcrypt.ErrMismatchedHashAndPassword {
			return PostLogin401JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeInvalidEmailOrPassword)
		}
		return PostLogin500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeInternalServer)
	}

	// create new user session
	ginCtx, _ := ctx.(*gin.Context)
	ip := ginCtx.ClientIP()
	ua := ginCtx.Request.UserAgent()

	// generate jwt
	token, err := a.Jwt.GenerateJWT(jwt.CreateUserClaims(strconv.Itoa(userID), "", email, []string{}))
	if err != nil {
		return PostLogin500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeInternalServer)
	}

	// generate refresh token
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return PostLogin500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeInternalServer)
	}
	refreshExpiresAt := time.Now().Add(7 * 24 * time.Hour) // 7 days

	tx, err := a.Db.DB.BeginTx(ctx, nil)
	if err != nil {
		return PostLogin500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeDBTransaction)
	}
	defer tx.Rollback()

	// store session with refresh token
	if _, err = tx.ExecContext(ctx, `
	INSERT INTO user_sessions (user_id, access_token, refresh_token, user_agent, ip_address, expires_at)
	VALUES ($1, $2, $3, $4, $5, $6);
`, strconv.Itoa(userID), token, refreshToken, ua, ip, refreshExpiresAt); err != nil {
		return PostLogin500JSONResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return PostLogin500JSONResponse{}, err
	}

	resp := LoginResponse{
		Data: &struct {
			Email        *string "json:\"email,omitempty\""
			RefreshToken *string "json:\"refresh_token,omitempty\""
			Token        *string "json:\"token,omitempty\""
		}{
			Email:        &request.Body.Email,
			RefreshToken: &refreshToken,
			Token:        &token,
		},
		Message:    "success login",
		StatusCode: http.StatusOK,
	}

	return PostLogin200JSONResponse(resp), nil
}

func (a *AuthImpl) PostRefresh(ctx context.Context, request PostRefreshRequestObject) (PostRefreshResponseObject, error) {

	refreshToken := request.Body.RefreshToken
	// check refresh token
	row := a.Db.DB.QueryRowContext(ctx, `
		SELECT us.user_id, us.user_agent, us.is_valid, us.expires_at, u.email
		FROM user_sessions us
		LEFT JOIN users u ON us.user_id = u.id
		WHERE refresh_token = $1;
	`, refreshToken)

	var userID string
	var userAgent string
	var isValid int
	var expiresAt time.Time
	var userEmail string
	if err := row.Scan(&userID, &userAgent, &isValid, &expiresAt, &userEmail); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PostRefresh401JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeInvalidRefreshToken)
		}
		return PostRefresh500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeDBQuery)
	}

	if isValid != 1 || time.Now().After(expiresAt) {
		return PostRefresh401JSONResponse{}, errlib.NewAppErrorWithLog(errors.New("expired or invalid refresh token"), errlib.ErrCodeInvalidRefreshToken)
	}

	// generate new jwt and rotate refresh token
	newToken, err := a.Jwt.GenerateJWT(jwt.CreateUserClaims(userID, userEmail, "", []string{}))
	if err != nil {
		return PostRefresh500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeInternalServer)
	}

	newRefresh, err := generateRefreshToken()
	if err != nil {
		return PostRefresh500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeInternalServer)
	}

	tx, err := a.Db.DB.BeginTx(ctx, nil)
	if err != nil {
		return PostRefresh500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeDBTransaction)
	}
	defer tx.Rollback()

	// invalidate old session
	// store session with refresh token
	if _, err = tx.ExecContext(ctx, `
		UPDATE user_sessions
		SET is_valid = 0
		WHERE refresh_token = $1;
	`, refreshToken); err != nil {
		return PostRefresh500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeDBTransaction)
	}

	ginCtx, _ := ctx.(*gin.Context)
	ip := ginCtx.ClientIP()

	// insert new session
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO user_sessions (user_id, access_token, refresh_token, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, newToken, newRefresh, userAgent, ip, time.Now().Add(7*24*time.Hour)); err != nil {
		return PostRefresh500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeDBTransaction)
	}

	if err := tx.Commit(); err != nil {
		return PostRefresh500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeDBTransaction)
	}
	// return

	resp := RefreshResponse{
		Message:      "token refreshed",
		StatusCode:   http.StatusOK,
		Token:        &newToken,
		RefreshToken: &newRefresh,
	}

	return PostRefresh200JSONResponse(resp), nil
}

func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
