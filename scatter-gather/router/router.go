package main

import (
	"encoding/json"
	"errors"
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

type shardResponse struct {
	data map[string]string
	err  error
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

func (ro *Router) getShard(key string) string {
	shardIdx := hashKey(key) % uint32(len(ro.shards))
	shard := ro.shards[shardIdx]
	return shard
}

func (ro *Router) routeRequest(r *http.Request, url string) ([]byte, error) {
	req, err := http.NewRequest(r.Method, url, nil)
	if err != nil {
		return nil, errors.New("failed to forward request")
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("failed to forward request")
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed to read request body")
	}

	return body, nil

}

func (ro *Router) handleRequest(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")

	shardIdx := hashKey(key) % uint32(len(ro.shards))
	shard := ro.shards[shardIdx]

	url := shard + r.URL.Path + "?" + r.URL.RawQuery
	resp, err := ro.routeRequest(r, url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (ro *Router) handleGetAll(w http.ResponseWriter, r *http.Request) {

	respChan := make(chan shardResponse, len(ro.shards))
	for _, shard := range ro.shards {
		go func(r *http.Request, url string) {
			resp, err := ro.routeRequest(r, url)
			if err != nil {
				respChan <- shardResponse{err: err}
				return
			}

			var m map[string]string
			if err := json.Unmarshal(resp, &m); err != nil {
				respChan <- shardResponse{err: err}
				return
			}

			respChan <- shardResponse{data: m}
		}(r, shard+r.URL.Path)
	}

	merged := make(map[string]string)
	for i := 0; i < len(ro.shards); i++ {
		items := <-respChan
		if items.err != nil {
			continue
		}

		for k, v := range items.data {
			merged[k] = v
		}

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(merged)
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
	http.HandleFunc("GET /all", router.handleGetAll)

	log.Print("[router]: listening on port 8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}
