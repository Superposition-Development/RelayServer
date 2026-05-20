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
	Offset    string `json:"offset"`
	ChannelID string `json:"channelID"`
	ServerID  string `json:"serverID"`
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

	/*

	   queryOffset = data["offset"]
	   queryChannel = data["channel"]
	   dbMessageList = Message.query.filter_by(channel=queryChannel).order_by(db.desc(Message.id)).offset(queryOffset).limit(5).all()
	   messageList = []
	   for message in dbMessageList:
	       messageUser = Users.query.filter_by(id=message.user).first()
	       messageList.append({
	           'name':messageUser.username,
	           'pfp':messageUser.pfp,
	           'message':message.content,
	           'date':message.date,
	           'userID':messageUser.userID
	       })
	   channelName = Channel.query.filter_by(id=queryChannel).first().name
	   return jsonify({'messages': messageList,
	                   "channelName":channelName})

	*/
}
