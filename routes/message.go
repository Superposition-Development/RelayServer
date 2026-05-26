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

type GetChannelMessageRequest struct {
	MoreThan  string `json:"moreThan"`
	ChannelID string `json:"channelID"`
	ServerID  string `json:"serverID"`
	MessageID string `json:"messageID"`
	Ascending string `json:"ascending"`
}

func SendMessage(w http.ResponseWriter, r *http.Request) {

	var data SendMessageRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		fmt.Println(err)
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
	db.AddRowWithIDReturn(newMessage, "message")
}

func GetChannelMessage(w http.ResponseWriter, r *http.Request) {
	var data GetChannelMessageRequest
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

	messages, err := db.PaginatedQuery([]string{"userID"}, "message", "channelID", data.ChannelID, "id", data.MoreThan == "true", data.MessageID, data.Ascending == "true", 15)
	userContentMap := make(map[string]map[string]any)

	for _, value := range messages {
		var rawUserID interface{} = value["userID"]
		userID := fmt.Sprintf("%v", rawUserID) //cooked
		_, ok := userContentMap[userID]
		if ok {
			value["pfp"] = userContentMap["pfp"]
			value["name"] = userContentMap["string"]
		} else {
			userContentMap[userID] = make(map[string]any)
			//todo here: query the user table and grab the pfp and name data and set it here
		}
	}

	// servers, err := db.Query([]string{"id", "pfp", "name"}, "server", "id", reformattedServerIDs)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
