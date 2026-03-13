from init import socketApp
import json
from AuthKeyGen import decryptJWT
import database
from routes.server import getServerUsers, getServers
from flask_sock import ConnectionClosed
from simple_websocket import Server

connections = {}
"""
{
    ws: {"servers": [], userID:userID}
}
"""

def removeConnection(ws):
    del connections[ws]

@socketApp.route("/ws")
def websocket(ws : Server):
    try:
        while True: 
            data = json.loads(ws.receive())

            if not "message" in data or "authKey" not in data:
                continue

            result = decryptJWT(data["authKey"])

            if not ("Result" in result.keys()):
                continue

            response = {"message":""}
            print(connections)

            match data["message"]:
                case "sendMessage":
                    response["message"] = "recieveMessage"
                    rawServerUsers = getServerUsers(data["serverID"]) #if this is a problem get the serverID from the channelID
                    serverUsers = set([userElement[0] for userElement in rawServerUsers])
                    for websocket, userData in connections.items():
                        if(userData["userID"] in serverUsers):
                            websocket.send(json.dumps(response))
                case "register":
                    user = database.queryTableValue(["id","pfp","username","userID","password","timestamp"],"user","userID",result["Result"]["userID"])
                    connections[ws] = {
                    "servers": getServers(user, True),
                    "userID":user["userID"]
                    }
                case "_":
                    continue
    except ConnectionClosed:
        removeConnection(ws)
