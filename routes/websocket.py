from init import socketApp
import json
from AuthKeyGen import decryptJWT
import database

connections = {}

@socketApp.route("/ws")
def websocket(ws):
    while True: 
        data = json.loads(ws.receive())

        if not "message" in data or "authKey" not in data:
            continue

        result = decryptJWT(data["authKey"])

        if not ("Result" in result.keys()):
            continue

        user = database.queryTableValue(["id","pfp","username","userID","password","timestamp"],"user","userID",result["Result"]["userID"])

        if not(user["userID"] in connections.keys()): #this might cause a vuln eventually if you can find a way to change the value of the websocket
            connections[user["userID"]] = ws

        response = {"message":""}

        match data["message"]:
            case "sendMessage":
                ws.send(json.dumps(response))
            case "_":
                continue
        
