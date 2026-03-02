from init import socketApp

@socketApp.route("/ws")
def websocket(ws):
    while True:
        text = ws.receive()
        ws.send(text[::-1])