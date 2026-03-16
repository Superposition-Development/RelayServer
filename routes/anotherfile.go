package routes

import (
	"fmt"
	"net/http"
)

func RegisterTestEndpoints() {
	http.HandleFunc("/test", test)
}

func test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "woo")
}
