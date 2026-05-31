package main

import (
	db "RelayServer/database"
	routes "RelayServer/routes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	ID   string
	Conn *websocket.Conn
}

var (
	clients = make(map[string]*Client)
	mu      sync.RWMutex
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()
	var data map[string]string
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}

		json.Unmarshal(msg, &data)
		switch data["message"] {
		case "register":
			user, err := db.AuthValidation(data["authKey"])
			if err != nil {
				//do something
			}
			var rawUserID interface{} = user["userID"]
			userID := fmt.Sprintf("%v", rawUserID)
			mu.Lock()
			clients[userID] = &Client{
				ID:   userID,
				Conn: ws,
			}
			mu.Unlock()
		case "sendMessage":

		}
	}

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Let's put something cool here eventually")
}

func registerEndpoints() {
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/signup", routes.Signup)
	http.HandleFunc("/login", routes.Login)
	http.HandleFunc("/getServers", routes.GetServers)
	http.HandleFunc("/createServer", routes.CreateServer)
	http.HandleFunc("/createChannel", routes.CreateChannel)
	http.HandleFunc("/getChannels", routes.GetChannels)
	http.HandleFunc("/sendMessage", routes.SendMessage)
	http.HandleFunc("/getMessages", routes.GetChannelMessage)
	http.HandleFunc("/validateUserToken", routes.ValidateUserToken)
}

func main() {
	db.InitializeConfig()
	// database.TheTrucksAreHere()
	db.InitializeDB()
	registerEndpoints()

	fmt.Println("Relay Server active on port 8080")
	handler := enableCORS(http.DefaultServeMux)
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		//lowk what do we even do here
	}
}
