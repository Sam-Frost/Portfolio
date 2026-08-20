package main

import (
	"fmt"
	"net/http"
)

func healthCheck(w http.ResponseWriter, req *http.Request) {

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Server is healthy, boi!!!")
}

func main() {
	fmt.Println("Server is starting to listen on port 8080...")

	http.HandleFunc("/health", healthCheck)

	http.ListenAndServe(":8080", nil)
}
