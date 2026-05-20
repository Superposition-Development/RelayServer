package routes

import (
	db "RelayServer/database"
	"encoding/json"
	"fmt"
	"net/http"
)

type CreateChannelRequest struct {
	Name     string `json:"name"`
	ServerID string `json:"serverID"`
}

type GetChannelsRequest struct {
	ServerID string `json:"serverID"`
}

func CreateChannel(w http.ResponseWriter, r *http.Request) {
	var data CreateChannelRequest
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

	createdChannel := map[string]any{
		"name":     data.Name,
		"serverID": data.ServerID,
	}

	_, e := db.AddRowWithIDReturn(createdChannel, "channel")
	if e != nil {

	}
	//do some thing abt check the headers for the JWT and then create serverUser
	fmt.Fprintf(w, "created server")
}

func GetChannels(w http.ResponseWriter, r *http.Request) {
	var data GetChannelsRequest
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

	// channels, err := db.Query([]string{"name", "id"}, "channel", "serverID", data.ServerID)
	// channels.
	// fmt.Fprintf(w,channels)
	//this part do something with it

}
