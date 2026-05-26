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

	for _, message := range messages {
		var rawUserID interface{} = message["userID"]
		userID := fmt.Sprintf("%v", rawUserID) //cooked
		_, ok := userContentMap[userID]
		if ok {
			message["pfp"] = userContentMap[userID]["pfp"]
			message["name"] = userContentMap[userID]["name"]
		} else {
			userContentMap[userID] = make(map[string]any)
			queryMap := map[string]string{
				"userID": userID,
			}
			userData, err := db.QueryRow([]string{"pfp", "username"}, "user", queryMap)
			if err != nil {
				fmt.Println(err)
			}
			userContentMap[userID]["pfp"] = userData["pfp"]
			userContentMap[userID]["name"] = userData["name"]
			message["pfp"] = userData["pfp"]
			message["name"] = userData["name"]
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
