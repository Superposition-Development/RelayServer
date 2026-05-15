package routes

import (
	db "RelayServer/database"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CreateChannelRequest struct {
	Name     string `json:"name"`
	ServerID string `json:"serverID"`
}

func CreateChannel(w http.ResponseWriter, r *http.Request) {
	var data CreateChannelRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}
	user, err := db.AuthHeaderValidation(r)
	if err != nil {
		//do something
	}

	var rawUserID interface{} = user["userID"]
	userID := fmt.Sprintf("%v", rawUserID) //cooked

	if !UserInServer(data.ServerID, userID) {

	}

	serverData := map[string]any{
		"pfp":       data.Pfp,
		"name":      data.Name,
		"timestamp": time.Now().Unix(),
	}

	serverID, err := db.AddRowWithIDReturn(serverData, "server")
	if err != nil {
		//do something
	}
	//do some thing abt check the headers for the JWT and then create serverUser
	CreateServerUser(serverID, userID)
	fmt.Fprintf(w, "created server")
}
