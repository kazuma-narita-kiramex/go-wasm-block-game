package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	dir := "./src"
	if len(os.Args) >= 2 {
		dir = os.Args[1]
	}
	const addr = "localhost:8085"
	log.Printf("Run Web Server on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, http.FileServer(http.Dir(dir))))
}
