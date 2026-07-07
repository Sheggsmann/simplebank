package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

// JWTMaker is a JSON WEB TOKEN maker
type JWTMaker struct {
	secretKey string
}

func (j *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	jwtPayload, err := NewPayload(username, duration)
	if err != nil {
		return "", fmt.Errorf("error creating token %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtPayload)
	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTMaker) VerifyToken(token string) (*Payload, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &Payload{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	} else if claims, ok := parsedToken.Claims.(*Payload); ok {
		return claims, nil
	} else {
		return nil, ErrInvalidToken
	}
}

// NewJWTMaker creates a new JWTMaker
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey: secretKey}, nil
}
