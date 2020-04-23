package main

import (
	"log"
	"net/http"
	"os"

	"github.com/apex/gateway"
)

func main() {

	aws := len(os.Getenv("AWS_REGION")) > 0

	r := router{devMode: !aws}
	http.Handle("/", r.handler())

	if aws {
		log.Fatal(gateway.ListenAndServe(":3000", nil))
	} else {
		log.Println("Starting listener http://localhost:3000")
		log.Fatal(http.ListenAndServe(":3000", nil))
	}
}
