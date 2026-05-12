// from flask import Blueprint, request, jsonify
// import database
// from AuthKeyGen import requiresToken
// import time

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
	//do some thing abt check the headers for the JWT and then create serverUser
	CreateServerUser(serverID, userID)
	fmt.Fprintf(w, "created server")
}

func userInServer() {}

// #TODO: make this secure / find some way to be like invitation link if it doesn't match this uuid
// @bpServer.route("/joinServer",methods=["POST"])
// @requiresToken
// def joinServer(user):
//     data = request.get_json()
//     createServerUser(data["serverID"],user["userID"])
//     response = jsonify({
//             "Message":"Joined server"
//         })
//     return response

func JoinServer(w http.ResponseWriter, r *http.Request) {
	var data JoinServerRequest
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

	// serverData := map[string]any{
	// 	"pfp":       data.Pfp,
	// 	"name":      data.Name,
	// 	"timestamp": time.Now().Unix(),
	// }

	// serverID, err := db.AddRowWithIDReturn(serverData, "server")
	// if err != nil {
	// 	//do something
	// }
	// //do some thing abt check the headers for the JWT and then create serverUser
	CreateServerUser(data.ServerID, userID)
	fmt.Fprintf(w, "created server")
}

// @bpServer.route("/createServer",methods=["POST"])
// @requiresToken
//     data = request.get_json()

//     createdServer = {
//         "name": data["name"],
//         "pfp": data["pfp"],
//         "timestamp": int(time.time())
//     }
//     serverID = database.addRowAndReturnRowID(createdServer,"server")
//     createServerUser(serverID,user["userID"])

//     response = jsonify({
//             "Message":"Server Created Successfully"
//         })
//     return response

// def userInServer(userID,serverID):
//     servers = database.queryTableValue("serverID","serverUser","userID",userID,True)
//     return any(row[0] == serverID for row in servers)

// #get all userIDs in a server
// def getServerUsers(serverID):
//     return database.queryTableValue("userID","serverUser","serverID",serverID,True)

// #TODO: do this for me please with database.py to delte elemets
// @bpServer.route("/leaveServer",methods=["POST"])
// @requiresToken
// def leaveServer(user):
//     pass

// def getServers(user,getForWebsocket):

//     servers = database.queryTableValue("serverID","serverUser","userID",user["userID"],True)
//     if(not servers):
//         return "No Servers Found"

//     serverList = []

//     for i in servers:
//         serverQuery = database.queryTableValue(["name","pfp"],"server","id",i[0])
//         if(getForWebsocket):
//             serverList.append(i)
//         else:
//             serverList.append({"name":serverQuery["name"],"pfp":serverQuery["pfp"],"id":i})

//         return serverList

// # get servers a user is in, i admit this is not a great endpoint name but i dont want a roblox looking method name thats 80 words long
// @bpServer.route("/getServers",methods=["GET"])
// @requiresToken
// def getServersWrapper(user,getForWebsocket=False):
//     result = getServers(user,getForWebsocket)
//     if(not getForWebsocket):
//         return jsonify({
//             "Message":result
//         })
//     return result
