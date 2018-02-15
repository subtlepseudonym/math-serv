package main

import (
	"log"
	"net/http"
	"time"

	"math-serv/server"
)

const defaultHost string = "127.0.0.1"
const defaultPort string = ":8080"

const defaultReadTimeout time.Duration = time.Second * 10
const defaultWriteTimeout time.Duration = time.Second * 10
const defaultIdleTimeout time.Duration = time.Second * 60

func main() {
	// setting server to default values, keeps door open to adding flags later
	srv := &http.Server{
		Addr:         defaultHost + defaultPort,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
		Handler:      server.GetRouter(),
	}

	log.Printf("Listening on %s\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
