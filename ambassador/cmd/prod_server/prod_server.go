package main

import (
	"fmt"
	"log"
	"net/http"
)

func prodHandlerFunc(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "this is the prod server!")

}
func main() {
	http.HandleFunc("/some-resource", prodHandlerFunc)

	log.Printf("[Prod Server]: Running at Port 8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
}
