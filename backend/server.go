package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

func main() {
	relativePath := "./wikipedia"
	absolutePath, err := filepath.Abs(relativePath)
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}

	initializeIndex(absolutePath)

	startTime := time.Now()
	calculatePageRank()
	elapsedTime := time.Since(startTime)
	fmt.Printf("PageRank: %.2fs\n", elapsedTime.Seconds())

	fmt.Println("Listening on port: 8080")

	startServer()
}

// init HTTP server
func startServer() {
	http.HandleFunc("/search", searchHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// HTTP search handler
func searchHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		Word string `json:"word"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestData); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	query := requestData.Word
	results := performSearch(query)
	jsonResponse, err := json.Marshal(results)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}