package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
)

const version = "0.1"

var (
	hostPort string
)

func main() {
	flag.StringVar(&hostPort, "listen-at", "localhost:8080", "listen for HTTP requests at host:port")
	flag.Parse()

	// Why we need "/" handler for the simple service? Beter to show
	// version on requests to root page for understanding that service
	// you have on this port. Getting the root page could be used by
	// monitoring service for health checks for example.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Sample Image Resizer ver.", version)
	})
	// Move real handling to another function for keeping main() short
	// and clean.
	http.HandleFunc("/resize", func(w http.ResponseWriter, r *http.Request) {
		handleResizeRequest(w, r)
	})
	if err := http.ListenAndServe(hostPort, nil); err != nil {
		panic(err)
	}

	terminate := make(chan os.Signal)
	signal.Notify(terminate, os.Interrupt, os.Kill)
	<-terminate
}
