package main

import (
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"time"
)

// i'm not adding a mutex here
// because i barely even access the shards (except during initialization)
type Router struct {
	shards []string
}

func newRouter(shards []string) *Router {
	return &Router{
		shards: shards,
	}
}

func hashKey(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (ro *Router) handleRequest(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	shardIdx := hashKey(key) % uint32(len(ro.shards))
	shard := ro.shards[shardIdx]

	url := shard + r.URL.Path + "?" + r.URL.RawQuery
	req, err := http.NewRequest(r.Method, url, nil)
	if err != nil {
		http.Error(w, "failed to forward request", http.StatusInternalServerError)
		return
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "failed to forward request", http.StatusInternalServerError)
		return

	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

func main() {

	// i'm just hard coding the number of shards
	// and their respective addresses. ideally these should
	// be dynamic
	numShards := 3
	port := 8080

	shards := make([]string, 0, numShards)
	for i := range numShards {
		shard := fmt.Sprintf("http://localhost:%d", port+i)
		shards = append(shards, shard)

	}

	router := newRouter(shards)
	http.HandleFunc("GET /get", router.handleRequest)
	http.HandleFunc("PUT /set", router.handleRequest)

	log.Print("[router]: listening on port 8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}
