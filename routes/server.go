package routes

import (
	db "RelayServer/database"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CreateServerRequest struct {
	Name string `json:"name"`
	Pfp  string `json:"pfp"`
}

type JoinServerRequest struct {
	ServerID string `json:"serverID"`
}

func CreateServerUser(serverID string, userID string) {

	serverUser := map[string]any{
		"serverID":  serverID,
		"userID":    userID,
		"timestamp": time.Now().Unix(),
	}
	db.AddRowWithIDReturn(serverUser, "serverUser")
}

func CreateServer(w http.ResponseWriter, r *http.Request) {
	var data CreateServerRequest
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

	serverData := map[string]any{
		"pfp":       data.Pfp,
		"name":      data.Name,
		"timestamp": time.Now().Unix(),
	}

	serverID, err := db.AddRowWithIDReturn(serverData, "server")
	if err != nil {
		//do something
	}

	CreateServerUser(serverID, userID)
	fmt.Fprintf(w, "created server")
}

func UserInServer(serverID string, userID string) bool {
	queryMap := map[string]string{
		"userID": userID,
	}
	servers, err := db.QueryRow([]string{"serverID"}, "serverUser", queryMap)
	fmt.Println(err)
	return len(servers) == 0
}

func GetServerUsers(serverID string) ([]string, error) {
	userMap, err := db.Query([]string{"userID"}, "serverUser", "serverID", []string{serverID})
	if err != nil {
		return nil, err
	}
	var userIDs []string
	for _, value := range userMap {
		userIDs = append(userIDs, fmt.Sprintf("%v", value["userID"]))
	}
	return userIDs, nil
}

func JoinServer(w http.ResponseWriter, r *http.Request) {
	var data JoinServerRequest
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}
	user, err := db.AuthHeaderValidation(r)
	if err != nil {
		fmt.Println(err)
	}

	var rawUserID interface{} = user["userID"]
	userID := fmt.Sprintf("%v", rawUserID) //cooked

	CreateServerUser(data.ServerID, userID)
	fmt.Fprintf(w, "created server")

	SendWebsocketMessage(clients[userID].Conn, WebsocketMessage{
		Type: "newServer",
		Data: map[string]any{},
	})
}

func GetServers(w http.ResponseWriter, r *http.Request) {
	user, err := db.AuthHeaderValidation(r)
	if err != nil {
		//do something
	}

	var rawUserID interface{} = user["userID"]
	userID := fmt.Sprintf("%v", rawUserID) //cooked

	serverIDs, err := db.Query([]string{"serverID"}, "serverUser", "userID", []string{userID})
	if err != nil {
		fmt.Println(err)
	}

	reformattedServerIDs := []string{}

	for i := 0; i < len(serverIDs); i++ {
		var serverid interface{} = serverIDs[i]["serverID"]
		reformattedServerIDs = append(reformattedServerIDs, fmt.Sprintf("%v", serverid))
	}

	servers, err := db.Query([]string{"id", "pfp", "name"}, "server", "id", reformattedServerIDs)

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
