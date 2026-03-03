from init import socketApp
from flask import jsonify
import json

connections = {}
print("breh")

@socketApp.route("/ws")
def websocket(ws):
    while True: 
        print("breh")
        data = json.loads(ws.receive())

        if not "message" in data:
            continue

        response = {"message":""}

        match data["message"]:
            case "sendMessage":
                ws.send(json.dumps(response))
            case "_":
                continue
        
