package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	fs := http.FileServer(http.Dir("./frontend"))
	frontendPort := os.Getenv("FRONTEND_PORT")
	if frontendPort == "" {
		frontendPort = "3000"
	}
	log.Printf("Frontend server started at http://localhost:%s\n", frontendPort)
	http.Handle("/", fs)
	log.Fatal(http.ListenAndServe(":"+frontendPort, nil))
}
