package helper

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
)

func LoadRSAPrivateKey(s string) (*rsa.PrivateKey, error) {
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

func LoadRSAPublicKey(s string) (*rsa.PublicKey, error) {
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
