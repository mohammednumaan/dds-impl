package main

import (
	"flag"
	"io"
	"log"
	"math/rand"
	"net/http"
)

var (
	prodServerAddr string
	betaServerAddr string
	splitPercent   int
)

// here, i need to split requests such that
// x% of requests go to the beta server and the rest to the prod server.
func splitRequestHandler(w http.ResponseWriter, req *http.Request) {
	target := prodServerAddr
	if rand.Intn(100) < splitPercent {
		target = betaServerAddr
	}

	// this is basically forwarding the request to the target server
	outboundReq, err := http.NewRequest(req.Method, target+req.URL.Path, req.Body)
	if err != nil {
		http.Error(w, "failed to build request", http.StatusInternalServerError)
		return
	}

	outboundReq.Header = req.Header.Clone()
	resp, err := http.DefaultClient.Do(outboundReq)
	if err != nil {
		http.Error(w, "failed to forward request", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	for k, v := range resp.Header {
		w.Header()[k] = v
	}

	// finally, i copy the response of the target server
	// to the writer so the client can get a response
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

}

func main() {
	flag.StringVar(&prodServerAddr, "prod", "http://localhost:8090", "prod server address")
	flag.StringVar(&betaServerAddr, "beta", "http://localhost:8091", "beta server address")
	flag.IntVar(&splitPercent, "nsplit", 10, "request split percentage")
	flag.Parse()

	http.HandleFunc("/some-resource", splitRequestHandler)
	log.Printf("[Ambassador Server]: Running at Port 8092")
	log.Fatal(http.ListenAndServe(":8092", nil))
}
