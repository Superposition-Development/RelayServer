from init import socketApp
import json
from AuthKeyGen import decryptJWT
import database
from routes.server import getServerUsers, getServers
from flask_sock import ConnectionClosed

connections = {}

"""
{
 userID:{"servers":[],
         "ws":ws}
}
"""

def updateConnectionServerList(userID,serverList):
    connections[userID]["servers"] = serverList

def removeConnection(userID):
    del connections[userID]    

print("bad")
@socketApp.route("/ws")
def websocket(ws):
    try:
        while True: 
            data = json.loads(ws.receive())

            if not "message" in data or "authKey" not in data:
                continue

            result = decryptJWT(data["authKey"])

            if not ("Result" in result.keys()):
                continue

            user = database.queryTableValue(["id","pfp","username","userID","password","timestamp"],"user","userID",result["Result"]["userID"])

            if not(user["userID"] in connections.keys()): #this might cause a vuln eventually if you can find a way to change the value of the websocket
                connections[user["userID"]] = {"servers":getServers(user,True), "ws":ws}

            response = {"message":""}

            match data["message"]:
                case "sendMessage":
                    response["message"] = "recieveMessage"
                    rawServerUsers = getServerUsers(data["serverID"]) #if this is a problem get the serverID from the channelID
                    serverUsers = [userElement[0] for userElement in rawServerUsers]
                    onlineUsers = list(set(serverUsers) & set(connections.keys()))
                    for socket in onlineUsers:
                        connections[socket].send(json.dumps(response))
                case "_":
                    continue
    except ConnectionClosed:
        pass