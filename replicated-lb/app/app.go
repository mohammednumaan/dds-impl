package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func handleApi(port int) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "[/api]: received request at :%d", port)

	}
}

func main() {
	portPtr := flag.Int("port", 8080, "port number to listen to http requests.")
	flag.Parse()

	port := *portPtr
	log.Printf("[server-%d]: listening at port, :%d", port, port)
	http.HandleFunc("/api", handleApi(port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
