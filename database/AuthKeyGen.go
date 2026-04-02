package database

import (
	"errors"
	"fmt"

	jwt "github.com/golang-jwt/jwt/v5"
)

type MyClaims struct {
	jwt.RegisteredClaims
	Username string `json:"username"`
}

func decryptJWT(tokenString string) (*jwt.Token, error) {
	secretKey := []byte(config.SecretKey)

	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		fmt.Println("Error parsing token:", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		fmt.Println("Token valid for user:", claims.Username)
	} else {
		fmt.Println("Invalid token")
	}
	return token, nil
}

func encryptJWT() {

}
