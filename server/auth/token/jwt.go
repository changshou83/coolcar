package token

import (
	"crypto/rsa"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWTTokenGen struct {
	privateKey *rsa.PrivateKey
	issuer     string
	nowFunc    func() time.Time
}

func NewJWTTokenGen(issuer string, privateKey *rsa.PrivateKey) *JWTTokenGen {
	return &JWTTokenGen{
		privateKey: privateKey,
		issuer:     issuer,
		nowFunc:    time.Now,
	}
}

func (gen *JWTTokenGen) GenerateToken(accountID string, expire time.Duration) (string, error) {
	nowSec := gen.nowFunc().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.StandardClaims{
		Issuer:    gen.issuer,                       // iss 谁发的
		IssuedAt:  nowSec,                           // iat 什么时候发的
		ExpiresAt: nowSec + int64(expire.Seconds()), // exp 什么时候过期
		Subject:   accountID,                        // sub 给谁发的
	})

	return token.SignedString(gen.privateKey)
}
