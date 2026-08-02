package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

/*
the config server is the SOURCE OF TRUTH for the .json or any
configuration files. any changes MUST be synced to the shared volume
that the app + sidecar container has access to
*/

type ConfigResponse struct {
	Hash    string `json:"hash"`
	Content string `json:"content"`
}

func main() {
	path := flag.String("file", "config.json", "path to config file to serve")
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Received request for config file: %s\n", *path)
		data, err := os.ReadFile(*path)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read config file: %v", err), http.StatusInternalServerError)
			return
		}

		sum := sha256.Sum256(data)
		hash := hex.EncodeToString(sum[:])

		response := ConfigResponse{
			Hash:    hash,
			Content: string(data),
		}

		w.Header().Set("Content-Type", "application/json")
		if json.NewEncoder(w).Encode(response) != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	})

	log.Printf("Starting config server on %s, serving file %s", *addr, *path)
	http.ListenAndServe(*addr, nil)
}
