package database

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type TokenClaims struct {
	jwt.RegisteredClaims
	UserID string `json:"userID"`
}

func DecryptJWT(tokenString string) (*jwt.Token, error) {
	secretKey := []byte(config.SecretKey)

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		fmt.Println("Error parsing token:", err)
		return nil, err
	}

	// if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
	// 	fmt.Println("Token valid for user:", claims.UserID)
	// } else {
	// 	fmt.Println("Invalid token")
	// }
	return token, nil
}

func EncryptJWT(userID string, expireTimeMin int) (string, error) {
	claim := TokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireTimeMin) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString([]byte(config.SecretKey))
}

func AuthHeaderValidation(r *http.Request) (any, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false, nil
	}
	token, err := DecryptJWT(strings.TrimPrefix(authHeader, "Bearer: "))
	if err != nil {
		//do something
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return QueryRow([]string{"id", "pfp", "username", "userID", "password", "timestamp"}, "user", "userID", claims.UserID)
	} else {
		fmt.Println("Invalid token")
	}
	return false, nil
}

func HashPassword(password string) {
	bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}
