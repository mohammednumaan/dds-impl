package main

import (
	"fmt"
	"log"
	"net/http"
)

func betaHandlerFunc(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "this is the beta server!")

}
func main() {
	http.HandleFunc("/some-resource", betaHandlerFunc)
	log.Printf("[Beta Server]: Running at Port 8091")
	log.Fatal(http.ListenAndServe(":8091", nil))
}
