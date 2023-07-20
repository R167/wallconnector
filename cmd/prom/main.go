// Pull stats from a wall connector and serve them as prometheus metrics.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	addr   = flag.String("addr", "localhost:8080", "address to listen on")
	path   = flag.String("path", "/metrics", "path to serve metrics on")
	target = flag.String("target", "localhost:8081", "target to forward requests to")
)

func main() {
	// Listen on the specified address and serve prometheus metrics
	// from the wall connector target.
	flag.Parse()

	// Create a new client for the wall connector.
	client, err := NewClient(*target)
	if err != nil {
		panic(err)
	}

	// Create a new collector for the wall connector.
	collector := NewCollector(client)

	// Register the collector with the prometheus default registry.
	http.Handle(*path, promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
