package routes

import (
	db "RelayServer/database"
	"encoding/json"
	"net/http"
)

type SignupRequest struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	UserID         string `json:"userID"`
	Pfp            string `json:"pfp"`
	SignupPassword string `json:"signupPassword"`
}

type SignupResponse struct {
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
	row, err := db.QueryRow([]string{"id"}, "user", "userID", "user")
	if len(row) != 0 {
		http.Error(w, "User Already Exists", http.StatusUnauthorized)
	}

	// if()
	// db.QueryRow()
}

func Login(w http.ResponseWriter, r *http.Request) {
	var data SignupRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}
	row, err := db.QueryRow([]string{"id"}, "user", "userID", "user")
	if len(row) != 0 {
		http.Error(w, "User Already Exists", http.StatusUnauthorized)
	}

}

// from flask import Blueprint, request, jsonify
// import database
// import AuthKeyGen
// from werkzeug.security import check_password_hash,generate_password_hash
// import time
// from init import SIGNUP_PASSWORD_REQUIRED, SIGNUP_PASSWORD

// bpAuth = Blueprint("auth",__name__)

// #TODO: sanitize userID
// @bpAuth.route("/signup",methods=["POST"])
// def signup():
//     data = request.get_json()

//     if(SIGNUP_PASSWORD_REQUIRED):
//         if(data["signupPassword"] != SIGNUP_PASSWORD):
//             return jsonify({"Error":"Invalid Credentials"})

//     query = database.queryTableValue("id","user","userID",data["userID"])

//     if(query != None):
//          return jsonify({"Error":"UserID taken"})

//     createdUser = {
//         "username": data["username"],
//         "password": generate_password_hash(data['password'], method='pbkdf2:sha256'),
//         "userID": data["userID"],
//         "pfp": data["pfp"],
//         "timestamp": int(time.time())
//     }

//     database.addRowAndReturnRowID(createdUser,"user")
//     key = AuthKeyGen.encryptJWT({"userID":data["userID"]},60)
//     response = jsonify({
//             "RelayJWT":key,
//             "userID":data["userID"]
//         })
//     return response

// @bpAuth.route("/login",methods=["POST"])
// def login():
//     data = request.get_json()
//     userQuery = database.queryTableValue(["password","userID"],"user","userID",data["userID"])
//     if(userQuery == None):
//         return jsonify({"Error":"Invalid Credentials"}) #sure they could just check with /signup to scan for userIDs, but wtv, hopefully anti brute force is written
//     if(not(check_password_hash(userQuery["password"],data["password"])) or userQuery["userID"] != data["userID"]):
//         return jsonify({"Error":"Invalid Credentials"})

//     key = AuthKeyGen.encryptJWT({"userID":data["userID"]},60)
//     response = jsonify({
//             "RelayJWT":key,
//             "userID":data["userID"]
//         })
//     return response
