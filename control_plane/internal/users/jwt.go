package users

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// tokenTTL is how long an issued token remains valid before the caller must
// log in again.
const tokenTTL = 24 * time.Hour

// ErrInvalidToken is returned by Verify for a missing, malformed, expired,
// or badly-signed token.
var ErrInvalidToken = errors.New("invalid token")

// Claims is the JWT payload: which user, which role.
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Sign issues a signed token for user, valid for tokenTTL.
func Sign(secret string, user User) (string, error) {
	now := time.Now()
	claims := Claims{
		Username: user.Username,
		Role:     user.RoleName,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Verify parses and validates token, returning its claims.
func Verify(secret, tokenString string) (Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return Claims{}, ErrInvalidToken
	}
	return claims, nil
}
