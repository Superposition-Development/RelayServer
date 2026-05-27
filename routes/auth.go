package routes

import (
	db "RelayServer/database"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SignupRequest struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	UserID         string `json:"userID"`
	Pfp            string `json:"pfp"`
	SignupPassword string `json:"signupPassword"`
}

type LoginRequest struct {
	Password string `json:"password"`
	UserID   string `json:"userID"`
}

type Response struct {
	RelayJWT string `json:"RelayJWT"`
	UserID   string `json:"userID"`
}

func Signup(w http.ResponseWriter, r *http.Request) {
	var data SignupRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}
	queryMap := map[string]string{
		"userID": data.UserID,
	}
	row, err := db.QueryRow([]string{"id"}, "user", queryMap)
	if len(row) != 0 {
		http.Error(w, "User Already Exists", http.StatusUnauthorized)
		return
	}

	userData := map[string]any{
		"pfp":       data.Pfp,
		"username":  data.Username,
		"userID":    data.UserID,
		"password":  data.Password, //need to encrypt this
		"timestamp": time.Now().Unix(),
	}

	db.AddRowWithIDReturn(userData, "user")

	w.Header().Set("Content-Type", "application/json")
	jwt, err := db.EncryptJWT(data.UserID, 60)
	if err != nil {
		fmt.Print(err)
	}
	signup := Response{RelayJWT: jwt, UserID: data.UserID}
	response, err := json.Marshal(signup)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

// this function is an endpoint is only meant to be used on boot to determine if the user is even logged in on startup
func ValidateUserToken(w http.ResponseWriter, r *http.Request) {

	_, err := db.AuthHeaderValidation(r)
	status := "ok"
	if err != nil {
		status = "fail"
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]any{
		"status": status,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var data LoginRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}

	queryMap := map[string]string{
		"userID":   data.UserID,
		"password": data.Password,
	}
	row, err := db.QueryRow([]string{"userID", "password"}, "user", queryMap)
	if len(row) == 0 {
		http.Error(w, "No User Found", http.StatusUnauthorized)
	}

	w.Header().Set("Content-Type", "application/json")
	jwt, err := db.EncryptJWT(data.UserID, 60)

	signup := Response{RelayJWT: jwt, UserID: data.UserID}
	response, err := json.Marshal(signup)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(response)
}
