package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Request struct {
	Text string `json:"text"`
}

type Response struct {
	Result string `json:"result"`
}

func main() {
	http.HandleFunc("/api/toupper", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req Request
		json.NewDecoder(r.Body).Decode(&req)
		json.NewEncoder(w).Encode(Response{Result: strings.ToUpper(req.Text)})
	})

	println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}