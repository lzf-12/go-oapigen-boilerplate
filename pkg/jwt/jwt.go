package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"errors"
	"oapi-to-rest/pkg/helper"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	RS256 = "RS256"
)

type JwtConfig struct {
	PrivateKeyBase64 string
	PublicKeyBase64  string
	ExpiresInSecond  time.Duration
}

type TokenManager struct {
	alg             string
	privateKey      any
	publicKey       any
	ExpiresInSecond time.Duration
}

func NewRSAJwtInit(config *JwtConfig) (*TokenManager, error) {
	privKey, err := helper.LoadRSAPrivateKey(config.PrivateKeyBase64)
	if err != nil {
		return nil, err
	}
	pubKey, err := helper.LoadRSAPublicKey(config.PublicKeyBase64)
	if err != nil {
		return nil, err
	}
	return &TokenManager{
		alg:             RS256,
		privateKey:      privKey,
		publicKey:       pubKey,
		ExpiresInSecond: config.ExpiresInSecond,
	}, nil
}

func (tm *TokenManager) GenerateJWT(claims jwt.MapClaims) (string, error) {

	claims["exp"] = time.Now().Add(tm.ExpiresInSecond).Unix()
	token := jwt.NewWithClaims(jwt.GetSigningMethod(tm.alg), claims)

	switch tm.alg {
	case RS256:
		return token.SignedString(tm.privateKey.(*rsa.PrivateKey))
	default:
		return "", errors.New("unsupported algorithm")
	}
}

func (tm *TokenManager) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// refresh
func (tm *TokenManager) RefreshJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return tm.publicKey, nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("cannot extract claims")
	}

	// remove old expiration
	delete(claims, "exp")
	return tm.GenerateJWT(claims)
}

// validate RSA jwt
func (tm *TokenManager) ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return tm.publicKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}

func CreateUserClaims(userID, username, email string, roles []string) jwt.MapClaims {
	return jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"email":    email,
		"roles":    roles,
		"iss":      "your-app",
		"sub":      userID,
		"aud":      "your-api",
		"iat":      time.Now().Unix(),
	}
}
