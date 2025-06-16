package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
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
	privKey, err := loadRSAPrivateKey(config.PrivateKeyBase64)
	if err != nil {
		return nil, err
	}
	pubKey, err := loadRSAPublicKey(config.PublicKeyBase64)
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

// --- Helper functions to load keys from base64 string in env

func loadRSAPrivateKey(s string) (*rsa.PrivateKey, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	block, _ := pem.Decode(decodedBytes)
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Println("error parse private key")
		return nil, err
	}

	pk, _ := priv.(*rsa.PrivateKey)
	return pk, nil
}

func loadRSAPublicKey(s string) (*rsa.PublicKey, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	block, _ := pem.Decode(decodedBytes)
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Println("error parse public key")
		return nil, err
	}
	return pub.(*rsa.PublicKey), nil
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
