from init import socketApp
import json

connections = {}

@socketApp.route("/websocket")
def websocket(ws):
    while True: 
        data = json.loads(ws.receive())
        if not "message" in data:
            continue
        match data["message"]:
            case "sendMessage":
                ws.send("message recieved")
            case "_":
                continue
        
