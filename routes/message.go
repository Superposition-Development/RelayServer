package routes

import (
	db "RelayServer/database"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SendMessageRequest struct {
	ServerID  string `json:"serverID"`
	ChannelID string `json:"channelID"`
	Content   string `json:"content"`
}

func SendMessage(w http.ResponseWriter, r *http.Request) {

	var data SendMessageRequest
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
		//FAH
	}

	newMessage := map[string]any{
		"channelID": data.ChannelID,
		"userID":    userID,
		"timestamp": time.Now().Unix(),
		"content":   data.Content,
	}
	db.AddRowWithIDReturn(newMessage, "serverUser")
}
