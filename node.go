package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func generateHash(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()
	hash := sha256.New()
	io.Copy(hash, file)
	return hex.EncodeToString(hash.Sum(nil))
}

func main() {
	port := flag.String("port", "8001", "Port to run the node on")
	dataDir := flag.String("data", "./data/node1", "Directory to store object files")
	flag.Parse()

	os.MkdirAll(*dataDir, 0755)
	mux := http.NewServeMux()

	mux.HandleFunc("PUT /objects/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		filePath := filepath.Join(*dataDir, id)
		hashPath := filePath + ".sha256"

		file, _ := os.Create(filePath)
		io.Copy(file, r.Body)
		file.Close()

		hashString := generateHash(filePath)
		os.WriteFile(hashPath, []byte(hashString), 0644)

		w.WriteHeader(http.StatusCreated)
		log.Printf("Stored object: %s (SHA256: %s...)", id, hashString[:8])
	})

	mux.HandleFunc("GET /objects/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		filePath := filepath.Join(*dataDir, id)
		hashPath := filePath + ".sha256"

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		expectedHashBytes, err := os.ReadFile(hashPath)
		if err != nil {
			http.ServeFile(w, r, filePath) 
			return
		}
		expectedHash := string(expectedHashBytes)
		actualHash := generateHash(filePath) 

		if actualHash != expectedHash {
			log.Printf("🚨 CORRUPTION DETECTED in %s! Expected %s, got %s", id, expectedHash[:8], actualHash[:8])
			http.Error(w, "Data corrupted", http.StatusInternalServerError)
			return
		}

		http.ServeFile(w, r, filePath)
		log.Printf("Retrieved and verified object: %s", id)
	})

	mux.HandleFunc("GET /objects/dummy_check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("Secure Storage Node starting on port %s, saving data to %s", *port, *dataDir)
	log.Fatal(http.ListenAndServe(":"+*port, mux))
}
