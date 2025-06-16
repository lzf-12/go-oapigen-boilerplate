package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"oapi-to-rest/pkg/db"
	"oapi-to-rest/pkg/errlib"

	"golang.org/x/crypto/bcrypt"
)

type AuthImpl struct {
	Db *db.SQLite
}

var _ StrictServerInterface = (*AuthImpl)(nil)

func (a *AuthImpl) PostRegister(ctx context.Context, request PostRegisterRequestObject) (PostRegisterResponseObject, error) {

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(request.Body.Password), bcrypt.DefaultCost)
	if err != nil {
		return PostRegister500JSONResponse{}, err
	}

	tx, err := a.Db.DB.BeginTx(ctx, nil)
	if err != nil {
		return PostRegister500JSONResponse{}, err
	}

	var userID int64
	result, err := tx.ExecContext(ctx, `
		INSERT INTO users (email, first_name, last_name)
		VALUES (?, ?, ?)`,
		request.Body.Email, request.Body.FirstName, request.Body.LastName,
	)
	if err != nil {
		tx.Rollback()
		return PostRegister500JSONResponse{}, err
	}

	userID, err = result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return PostRegister500JSONResponse{}, err
	}

	// insert into auth_credentials table
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO auth_credentials (user_id, provider, password_hash)
		VALUES (?, 'local', ?)`,
		userID, string(hashPassword),
	); err != nil {
		tx.Rollback()
		return PostRegister500JSONResponse{}, err
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		return PostRegister500JSONResponse{}, err
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
		return PostLogin500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeDBQuery)
	}

	// compare hash with bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(request.Body.Password)); err != nil {

		if err == bcrypt.ErrMismatchedHashAndPassword {
			return PostLogin401JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeInvalidEmailOrPassword)
		}
		return PostLogin500JSONResponse{}, errlib.NewAppErrorWithLog(err, errlib.ErrCodeInternalServer)
	}

	resp := LoginResponse{Data: &struct {
		Email *string "json:\"email,omitempty\""
	}{
		Email: &request.Body.Email,
	},
		Message:    "success login",
		StatusCode: http.StatusOK}

	return PostLogin200JSONResponse(resp), nil
}

// TODO
// refresh jwt endpoint
