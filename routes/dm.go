package routes

import (
	db "RelayServer/database"
	"encoding/json"
	"fmt"
	"net/http"
)

type CreateDMRequest struct {
	TargetID string `json:"targetID"`
}

func CreateDM(w http.ResponseWriter, r *http.Request) {
	var data CreateDMRequest
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

	createdChannel := map[string]any{
		"name":     nil,
		"serverID": nil,
		"type":     "text",
	}

	textID, e := db.AddRowWithIDReturn(createdChannel, "channel")
	if e != nil {

	}

	//DRY !! Grow up m8

	createdChannel = map[string]any{
		"name":     nil,
		"serverID": nil,
		"type":     "voice",
	}

	voiceID, err := db.AddRowWithIDReturn(createdChannel, "channel")
	if err != nil {

	}

	dmData := map[string]any{
		"textID":    textID,
		"voiceID":   voiceID,
		"userOneID": userID,
		"userTwoID": data.TargetID,
	}

	_, err = db.AddRowWithIDReturn(dmData, "dm")
	if err != nil {
		//do something
	}

	SendWebsocketMessage(clients[userID].Conn, WebsocketMessage{
		Type: "newChannel",
		Data: map[string]any{},
	})

	fmt.Fprintf(w, "created server")
}

// func UserInDM(dmID string, userID string) bool {
// 	queryMap := map[string]string{
// 		"userID": userID,
// 	}
// 	servers, err := db.QueryRow([]string{"dmID"}, "dm", queryMap)
// 	fmt.Println(err)
// 	return len(servers) == 0
// }

func GetDMUsers(dmID string) ([]string, error) {
	userMap, err := db.Query([]string{"userA", "userB"}, "dm", "id", []string{dmID})
	if err != nil {
		return nil, err
	}
	var userIDs []string
	for _, value := range userMap {
		userIDs = append(userIDs, fmt.Sprintf("%v", value["userID"]))
	}
	return userIDs, nil
}

func GetDMs(w http.ResponseWriter, r *http.Request) {
	user, err := db.AuthHeaderValidation(r)
	if err != nil {
		//do something
	}

	var rawUserID interface{} = user["userID"]
	userID := fmt.Sprintf("%v", rawUserID) //cooked

	dmIDs, err := db.Query([]string{"id", "userA", "userB"}, "dm", "userID", []string{userID})
	if err != nil {
		fmt.Println(err)
	}

	reformattedDMIDs := []string{}

	for i := 0; i < len(dmIDs); i++ {
		var dmID interface{} = dmIDs[i]["serverID"]
		reformattedDMIDs = append(reformattedDMIDs, fmt.Sprintf("%v", dmID))
	}

	servers, err := db.Query([]string{"id", "pfp", "name"}, "dmID", "id", reformattedDMIDs)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(servers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
