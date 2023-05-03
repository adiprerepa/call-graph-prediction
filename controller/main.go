package main

import (
	"fmt"
	"net/http"
)

var ip string
  
func main() {
    port := "8080"
    fmt.Printf("Starting server at port %s", port)
    ip = "10.244.0.13:80"
    http.HandleFunc("/getEndpoint", HelloServer)

    http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        fmt.Fprintf(w, ip)
        return
    }
    if r.Method == "POST" {
        if newIp := r.URL.Query().Get("ip"); newIp != "" {
            ip = newIp
        }
    }
}