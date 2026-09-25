package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

type NodeStatus struct {
	URL    string `json:"url"`
	Status string `json:"status"`
}

var (
	nodes = []string{
		"http://localhost:8001",
		"http://localhost:8002",
		"http://localhost:8003",
	}
	fileLedger []string
	mu         sync.Mutex
)

func checkHealth(nodeURL string) string {
	client := http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(nodeURL + "/objects/dummy_check")
	if err != nil {
		return "OFFLINE"
	}
	defer resp.Body.Close()
	return "ONLINE"
}

// secureHeaders adds standard security headers to all responses
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("./web")))

	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
		status := struct { Nodes []NodeStatus `json:"nodes"` }{}
		for _, node := range nodes {
			status.Nodes = append(status.Nodes, NodeStatus{URL: node, Status: checkHealth(node)})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	mux.HandleFunc("GET /api/files", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"files": fileLedger})
	})

	mux.HandleFunc("PUT /objects/{id}", func(w http.ResponseWriter, r *http.Request) {
		// SECURITY: Prevent path traversal attacks
		id := filepath.Base(r.PathValue("id"))
		
		// SECURITY: Limit uploads to 10MB to prevent DoS
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "File too large or malformed", http.StatusBadRequest)
			return
		}
		
		successCount := 0

		for _, node := range nodes {
			url := fmt.Sprintf("%s/objects/%s", node, id)
			req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
			client := &http.Client{Timeout: 2 * time.Second}
			resp, reqErr := client.Do(req)
			if reqErr == nil && resp.StatusCode == http.StatusCreated {
				successCount++
			}
		}

		if successCount > 0 {
			mu.Lock()
			exists := false
			for _, f := range fileLedger { if f == id { exists = true } }
			if !exists { fileLedger = append(fileLedger, id) }
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
		} else {
			http.Error(w, "System failure", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("GET /objects/{id}", func(w http.ResponseWriter, r *http.Request) {
		// SECURITY: Prevent path traversal attacks
		id := filepath.Base(r.PathValue("id"))
		
		for _, node := range nodes {
			url := fmt.Sprintf("%s/objects/%s", node, id)
			client := &http.Client{Timeout: 1 * time.Second}
			resp, err := client.Get(url)
			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				io.Copy(w, resp.Body)
				return 
			}
		}
		http.Error(w, "Object not found or corrupted", http.StatusNotFound)
	})

	log.Println("Vault Control Plane Gateway running on port 8000...")
	// Wrap mux with security headers
	log.Fatal(http.ListenAndServe(":8000", secureHeaders(mux)))
}
