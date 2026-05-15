package main

import (
	database "RelayServer/database"
	routes "RelayServer/routes"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

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

func registerEndpoints() {
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/signup", routes.Signup)
	http.HandleFunc("/login", routes.Login)
	http.HandleFunc("/createServer", routes.CreateServer)
	// http.HandleFunc("")
}

func main() {
	database.InitializeConfig()
	// database.TheTrucksAreHere()
	database.InitializeDB()
	registerEndpoints()
	routes.UserInServer("1", "user")
	fmt.Println("Relay Server active on port 8080")
	handler := enableCORS(http.DefaultServeMux)
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		//lowk what do we even do here
	}
}
