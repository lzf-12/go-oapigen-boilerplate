package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"oapi-to-rest/pkg/db"

	"golang.org/x/crypto/bcrypt"
)

type AuthImpl struct {
	Db *db.SQLite
}

var _ StrictServerInterface = (*AuthImpl)(nil)

func (a *AuthImpl) PostRegister(ctx context.Context, request PostRegisterRequestObject) (PostRegisterResponseObject, error) {

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(request.Body.Password), bcrypt.DefaultCost)
	if err != nil {
		errRsp := ErrorResponse{
			Error:      fmt.Sprintf("error: %v", err),
			StatusCode: http.StatusInternalServerError,
			Trace:      "",
			Message:    "something went wrong during hashing",
		}
		return PostRegister500JSONResponse(errRsp), err
	}

	tx, err := a.Db.DB.BeginTx(ctx, nil)
	if err != nil {
		errRsp := ErrorResponse{
			Error:      fmt.Sprintf("error: %v", err),
			StatusCode: http.StatusInternalServerError,
			Trace:      "",
			Message:    "something went wrong during init db transaction",
		}
		return PostRegister500JSONResponse(errRsp), err
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
	_, err = tx.ExecContext(ctx, `
		INSERT INTO auth_credentials (user_id, provider, password_hash)
		VALUES (?, 'local', ?)`,
		userID, string(hashPassword),
	)
	if err != nil {
		tx.Rollback()
		errRsp := ErrorResponse{
			Error:      fmt.Sprintf("error: %v", err),
			StatusCode: http.StatusInternalServerError,
			Trace:      "",
			Message:    "something went wrong during db transaction",
		}
		return PostRegister500JSONResponse(errRsp), err
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		errRsp := ErrorResponse{
			Error:      fmt.Sprintf("error: %v", err),
			StatusCode: http.StatusInternalServerError,
			Message:    "something went wrong during db transaction",
		}
		return PostRegister500JSONResponse(errRsp), err
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

	var errRsp ErrorResponse

	// check email exist
	row := a.Db.DB.QueryRowContext(ctx, `
	SELECT id, email
	FROM users 
	WHERE email = $1 
	ORDER BY created_at DESC 
	LIMIT 1;
`, request.Body.Email)

	var userID int
	var email string
	var hashedPassword string
	if err := row.Scan(&userID, email, &hashedPassword); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errRsp.Message = "invalid email or password"
			errRsp.StatusCode = http.StatusUnauthorized
			errRsp.Error = "unauthorized"
			return PostLogin401JSONResponse(errRsp), nil
		}
		errRsp.Message = "invalid email or password"
		errRsp.StatusCode = http.StatusUnauthorized
		errRsp.Error = "unauthorized"
		return PostLogin500JSONResponse{}, err
	}

	// compare hash with bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(request.Body.Password)); err != nil {
		errRsp.Message = "invalid email or password"
		errRsp.StatusCode = http.StatusUnauthorized
		errRsp.Error = "unauthorized"
		return PostLogin401JSONResponse(errRsp), nil
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
