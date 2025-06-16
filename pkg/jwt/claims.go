package jwt

import "github.com/golang-jwt/jwt/v5"

type CustomClaims struct {
	UserID   string `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	jwt.RegisteredClaims
}
