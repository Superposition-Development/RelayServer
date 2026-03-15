package main

import (
	routes "RelayServer/routes"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("read error:", err)
			break
		}
		fmt.Printf("Received: %s\n", msg)

		if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
			fmt.Println("write error:", err)
			break
		}
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Let's put something cool here eventually")
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/", homeHandler)

	fmt.Println("WebSocket server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("ListenAndServe:", err)
	}
	routes.Test()
}
