package main

import (
	"fmt"
	"net/http"
)
  
func main() {
    http.HandleFunc("/getEndpoint", HelloServer)
    http.ListenAndServe(":80", nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "10.244.0.76:80")
}