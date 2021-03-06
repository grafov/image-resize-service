package main

import (
	"github.com/coocood/freecache"

	"flag"
	"net/http"
)

const version = "0.1"

const (
	// For production services the size setting obviously should be
	// move to config.
	cacheSize = 300 * 1024 * 1024
)

var (
	hostPort string
	cache    *freecache.Cache
)

func main() {
	flag.StringVar(&hostPort, "listen-at", "localhost:8080", "listen for HTTP requests at host:port")
	flag.Parse()

	cache = freecache.NewCache(cacheSize)

	// Why we need "/" handler for the simple service? Beter to show
	// version on requests to root page for understanding that service
	// you have on this port. Getting the root page could be used by
	// monitoring service for health checks for example.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRootRequest(w, r)
	})
	// Move real handling to another function for keeping main() short
	// and clean.
	http.HandleFunc("/resize", func(w http.ResponseWriter, r *http.Request) {
		handleResizeRequest(w, r)
	})
	if err := http.ListenAndServe(hostPort, nil); err != nil {
		panic(err)
	}
}
