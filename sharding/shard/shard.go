package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Shard struct {
	store map[string]string
	mu    sync.Mutex
	port  int
}

func newShard(port int) *Shard {
	return &Shard{
		store: make(map[string]string),
		port:  port,
	}
}

func (s *Shard) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	log.Printf("[shard-%d]: received GET key=%q from %s", s.port, key, r.RemoteAddr)
	if key == "" {
		http.Error(w, "missing 'key' parameter in request", http.StatusNotFound)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	val, ok := s.store[key]
	if !ok {
		http.Error(w, "could not find requested key", http.StatusNotFound)
		return
	}

	fmt.Fprint(w, val)
}

func (s *Shard) handleSet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	log.Printf("[shard-%d]: received PUT key=%q value=%q from %s", s.port, key, value, r.RemoteAddr)

	if key == "" {
		http.Error(w, "missing 'key' parameter in request", http.StatusNotFound)
		return
	}
	if value == "" {
		http.Error(w, "missing 'value' parameter in request", http.StatusNotFound)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[key] = value
	fmt.Fprintf(w, "successfully wrote k,v: {%s:%s}", key, value)
}

func main() {

	portPtr := flag.Int("port", 8080, "port to run the shard on")
	flag.Parse()
	portInt := *portPtr

	shard := newShard(portInt)
	http.HandleFunc("GET /get", shard.handleGet)
	http.HandleFunc("PUT /set", shard.handleSet)

	log.Printf("[shard-%d]: listening on port %d", portInt, portInt)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", portInt), nil))
}
