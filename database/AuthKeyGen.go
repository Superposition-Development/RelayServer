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
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
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

func AuthHeaderValidation(r *http.Request) (map[string]any, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("no auth or smth")
	}
	token, err := DecryptJWT(strings.TrimPrefix(authHeader, "Bearer "))
	if err != nil {
		//do something
		fmt.Print("FAH")
		fmt.Print(err)
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		queryMap := map[string]string{
			"userID": claims.UserID,
		}
		return QueryRow([]string{"id", "pfp", "username", "userID", "password", "timestamp"}, "user", queryMap)
	} else {
		return nil, errors.New("idk ts language way too hard")
	}
}

func HashPassword(password string) {
	bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}
