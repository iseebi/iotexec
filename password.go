package main

import (
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"time"
)

type Password interface {
	GetPassword() (string, error)
}

type RawPassword struct {
	password string
}

func NewRawPassword(password string) *RawPassword {
	return &RawPassword{password: password}
}

func (p RawPassword) GetPassword() (string, error) {
	return p.password, nil
}

type JwtPassword struct {
	privateKey *rsa.PrivateKey
	audience   string
}

func NewJwtPassword(privateKey *rsa.PrivateKey, audience string) *JwtPassword {
	return &JwtPassword{privateKey: privateKey, audience: audience}
}

func (p JwtPassword) GetPassword() (string, error) {
	now := time.Now()

	token := jwt.New(jwt.SigningMethodRS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["aud"] = p.audience
	claims["iat"] = now.Unix()
	claims["exp"] = now.Add(time.Minute * 20).Unix()

	return token.SignedString(p.privateKey)
}
