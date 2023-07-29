// Pull stats from a wall connector and serve them as prometheus metrics.
package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/R167/wallconnector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	addr   = flag.String("addr", "localhost:8080", "address to listen on")
	path   = flag.String("path", "/metrics", "path to serve metrics on")
	target = flag.String("target", "localhost:8081", "target to forward requests to")
)

func main() {
	start := time.Now()
	// Listen on the specified address and serve prometheus metrics
	// from the wall connector target.
	flag.Parse()

	// Create a new client for the wall connector.
	client, err := wallconnector.NewClient(*target)
	if err != nil {
		panic(err)
	}

	// Create a new collector for the wall connector.
	collector := wallconnector.NewCollector(client)

	reg := prometheus.NewRegistry()
	reg.MustRegister(collector)

	// Serve the metrics on the specified path.
	http.Handle(*path, promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ProcessStartTime: start,
	}))
	log.Printf("listening on %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		panic(err)
	}
}
