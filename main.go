package main

import (
	db "RelayServer/database"
	routes "RelayServer/routes"
	"fmt"
	"net/http"
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

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Let's put something cool here eventually")
}

func registerEndpoints() {
	http.HandleFunc("/ws", routes.HandleConnections)
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/signup", routes.Signup)
	http.HandleFunc("/login", routes.Login)
	http.HandleFunc("/getServers", routes.GetServers)
	http.HandleFunc("/createServer", routes.CreateServer)
	http.HandleFunc("/createChannel", routes.CreateChannel)
	http.HandleFunc("/getChannels", routes.GetChannels)
	http.HandleFunc("/getMessages", routes.GetChannelMessage)
	http.HandleFunc("/validateUserToken", routes.ValidateUserToken)
	http.HandleFunc("/joinServer", routes.JoinServer)
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
