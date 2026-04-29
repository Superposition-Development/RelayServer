// from flask import Blueprint, request, jsonify
// import database
// from AuthKeyGen import requiresToken
// import time

package routes

import (
	db "RelayServer/database"
	"time"
)

func CreateServerUser(serverID string, userID string) {

	serverUser := map[string]any{
		"serverID":  serverID,
		"userID":    userID,
		"timestamp": time.Now().Unix(),
	}
	db.AddRowWithIDReturn(serverUser, "serverUser")
}

// bpServer = Blueprint("server",__name__)

// def userInServer(userID,serverID):
//     servers = database.queryTableValue("serverID","serverUser","userID",userID,True)
//     return any(row[0] == serverID for row in servers)

// #get all userIDs in a server
// def getServerUsers(serverID):
//     return database.queryTableValue("userID","serverUser","serverID",serverID,True)

// def createServerUser(serverID,userID):
//     serverUser = {
//         "serverID": serverID,
//         "userID": userID,
//         "timestamp": int(time.time())
//     }
//     database.addRowAndReturnRowID(serverUser,"serverUser")

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

// #TODO: do this for me please with database.py to delte elemets
// @bpServer.route("/leaveServer",methods=["POST"])
// @requiresToken
// def leaveServer(user):
//     pass

// @bpServer.route("/createServer",methods=["POST"])
// @requiresToken
// def createServer(user):
//     """
//     Create a server in the server table

//     The expected payload from the client is as follows,

//     {
//         name: the name of the server to add
//         pfp: Base64 of picture
//     }

//     """
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
