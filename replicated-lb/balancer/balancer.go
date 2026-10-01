package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type Balancer struct {
	next     int
	replicas []string
	mu       sync.Mutex
}

func newBalancer(replicas []string) *Balancer {
	return &Balancer{
		next:     0,
		replicas: replicas,
	}
}

func (b *Balancer) getNextReplica() (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.replicas) == 0 {
		return "", false
	}

	replica := b.replicas[b.next]
	b.next = (b.next + 1) % len(b.replicas)

	return replica, true
}

func (b *Balancer) routeRequest(w http.ResponseWriter, req *http.Request) {
	replica, ok := b.getNextReplica()
	if !ok {
		http.Error(w, "no replicas available", http.StatusServiceUnavailable)
		return
	}
	client := &http.Client{
		Timeout: time.Second * 2,
	}

	resp, err := client.Get(replica + req.URL.Path)
	if err != nil {
		http.Error(w, "could not get response", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "could not read the response body", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(body)
}

func main() {

	numReplicas := 3
	replicas := make([]string, 0, numReplicas)
	currPort := 8080
	for i := 0; i < numReplicas; i++ {
		log.Printf("[balancer]: registered replica :%d", currPort+i)
		replicas = append(replicas, fmt.Sprintf("http://localhost:%d", currPort+i))
	}
	balancer := newBalancer(replicas)

	http.HandleFunc("/api", balancer.routeRequest)
	log.Fatal(http.ListenAndServe(":8090", nil))
}
