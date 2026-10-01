package routes

import (
	db "RelayServer/database"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// type SendMessageRequest struct {
// 	ServerID  string `json:"serverID"`
// 	ChannelID string `json:"channelID"`
// 	Content   string `json:"content"`
// }

type GetChannelMessageServerRequest struct {
	MoreThan  string `json:"moreThan"`
	ChannelID string `json:"channelID"`
	ServerID  string `json:"serverID"`
	MessageID string `json:"messageID"`
	Ascending string `json:"ascending"`
}

type GetChannelMessageDMRequest struct {
	MoreThan  string `json:"moreThan"`
	DMID      string `json:"dmID"`
	MessageID string `json:"messageID"`
	Ascending string `json:"ascending"`
}

func SendMessageServer(serverID string, channelID string, content string, JWT string) (string, error) {

	user, err := db.AuthValidation(JWT)
	if err != nil {
		//do something
	}

	var rawUserID interface{} = user["userID"]
	userID := fmt.Sprintf("%v", rawUserID) //cooked

	if !UserInServer(serverID, userID) {
		//FAH
	}

	newMessage := map[string]any{
		"channelID": channelID,
		"userID":    userID,
		"timestamp": time.Now().Unix(),
		"content":   content,
	}
	return db.AddRowWithIDReturn(newMessage, "message")
}

func SendMessageDM(dmID string, content string, JWT string) (string, error) {

	user, err := db.AuthValidation(JWT)
	if err != nil {
		//do something
	}

	var rawUserID interface{} = user["userID"]
	userID := fmt.Sprintf("%v", rawUserID) //cooked

	userInDM, err := db.UsersInDM(userID, dmID)

	if err != nil {
		//do smth about it
	}

	// fmt.Println(db.UsersInDM(useri))

	if userInDM == nil {
		return "no users this will break it but LT", nil
	}

	newMessage := map[string]any{
		"channelID": userInDM[3],
		"userID":    userID,
		"timestamp": time.Now().Unix(),
		"content":   content,
	}
	return db.AddRowWithIDReturn(newMessage, "message")
}

func GetChannelMessageServer(w http.ResponseWriter, r *http.Request) {
	var data GetChannelMessageServerRequest
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

	messages, err := db.PaginatedQuery([]string{"userID", "content", "id", "timestamp"}, "message", "channelID", data.ChannelID, "id", data.MoreThan == "true", data.MessageID, data.Ascending == "true", db.ServerConfig.MaxMessagesPerQuery)
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
				// fmt.Println(err)
			}
			userContentMap[userID]["pfp"] = userData["pfp"]
			userContentMap[userID]["name"] = userData["username"]
			message["pfp"] = userData["pfp"]
			message["name"] = userData["username"]
		}
	}

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

func GetChannelMessageDM(w http.ResponseWriter, r *http.Request) {
	var data GetChannelMessageDMRequest
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

	fmt.Println(data)

	users, err := db.UsersInDM(userID, data.DMID)
	if err != nil {
		fmt.Println(err)
	}
	if len(users) == 0 {
		// fmt.Println(users)
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// fmt.Println(users)
	// fmt.Println(users[3])

	messages, err := db.PaginatedQuery([]string{"userID", "content", "id", "timestamp"}, "message", "channelID", users[3], "id", data.MoreThan == "true", data.MessageID, data.Ascending == "true", db.ServerConfig.MaxMessagesPerQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
				// fmt.Println(err)
			}
			userContentMap[userID]["pfp"] = userData["pfp"]
			userContentMap[userID]["name"] = userData["username"]
			message["pfp"] = userData["pfp"]
			message["name"] = userData["username"]
		}
	}

	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(messages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
