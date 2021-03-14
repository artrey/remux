package main

import (
	"github.com/artrey/remux/pkg/remux"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	defaultHost = "0.0.0.0"
	defaultPort = "9999"
)

func main() {
	host, ok := os.LookupEnv("HOST")
	if !ok {
		host = defaultHost
	}
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = defaultPort
	}

	address := net.JoinHostPort(host, port)
	log.Println(address)

	if err := execute(address); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func execute(address string) error {
	mux := remux.New()
	server := http.Server{
		Addr:    address,
		Handler: mux,
	}
	return server.ListenAndServe()
}
