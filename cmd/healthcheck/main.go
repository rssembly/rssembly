// Package main provides a healthcheck binary for Docker/Kubernetes probes.
// Supports liveness (/health) and readiness (-readiness flag, /ready) checks.
// Reads SERVER_PORT from the environment to determine which port to check.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	readiness := flag.Bool("readiness", false, "check readiness (/ready) instead of liveness (/health)")
	flag.Parse()

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	path := "/health"
	if *readiness {
		path = "/ready"
	}

	url := "http://localhost:" + port + path
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "healthcheck: %s — %v\n", url, err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "healthcheck: %s — unexpected status %d\n", url, resp.StatusCode)
		os.Exit(1)
	}
}
