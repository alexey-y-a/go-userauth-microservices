package jwt

import (
	"errors"
	"os"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

const accessTokenTTL = time.Hour

var ErrInvalidToken = errors.New("invalid token")


type Claims struct {
    UserID string `json:"sub"`
    jwtlib.RegisteredClaims
}

func getSigningKey() ([]byte, error) {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        return nil, errors.New("JWT_SECRET is not set")
    }

    return []byte(secret), nil
}

func GenerateAccessToken(userID string) (string, error) {
    signingKey, err := getSigningKey()
    if err != nil {
        return "", err
    }

    now := time.Now().UTC()

    claims := Claims {
        UserID: userID,
        RegisteredClaims: jwtlib.RegisteredClaims{
            Issuer: "go-userauth-microservices",
            IssuedAt: jwtlib.NewNumericDate(now),
            ExpiresAt: jwtlib.NewNumericDate(now.Add(accessTokenTTL)),
            Subject: userID,
        },
    }

    token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)

    signedToken, err := token.SignedString(signingKey)
    if err != nil {
        return "", err
    }

    return signedToken, nil
}

func ParseToken(tokenString string) (Claims, error) {
    signingKey, err := getSigningKey()
    if err != nil {
        return Claims{}, err
    }

    claims := Claims{}

    token, err := jwtlib.ParseWithClaims(tokenString, &claims, func(token *jwtlib.Token) (interface{}, error) {
        _, ok := token.Method.(*jwtlib.SigningMethodHMAC)
        if !ok {
            return nil, ErrInvalidToken
        }
        return signingKey, nil
    })

    if err != nil {
        return Claims{}, ErrInvalidToken
    }

    if !token.Valid {
        return Claims{}, ErrInvalidToken
    }

    return claims, nil

}