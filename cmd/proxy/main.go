// Create a proxy server that will forward requests to the real server
// and return the response to the client.
//
// This is effectively a man in the middle proxy which allows us to log
// the contents of http requests and responses.
package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	addr   = flag.String("addr", "localhost:8080", "address to listen on")
	target = flag.String("target", "google.com:80", "target to forward requests to")
)

func main() {
	flag.Parse()

	log.Printf("listening on %s", *addr)
	if err := http.ListenAndServe(*addr, http.HandlerFunc(handler)); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Log the request
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("error dumping request: %v", err)
	}
	log.Printf("request:\n%s\n", dump)

	// Create a new request to forward to the real server
	// Note: we need to create a new request because the original
	// request has already been read and we can't read it again.
	// See: https://golang.org/pkg/net/http/#Request.Clone
	req := r.Clone(r.Context())

	// Update the request to point to the real server
	req.URL, err = url.Parse("http://" + *target + r.URL.Path)
	if err != nil {
		log.Printf("error parsing target url: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Host = *target
	req.RequestURI = ""

	// Forward the request to the real server
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("error forwarding request: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Log the response
	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		log.Printf("error dumping response: %v\n", err)
	}
	log.Printf("response:\n%s\n", dump)

	// Copy the response to the client
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("error copying response: %v", err)
	}
}
