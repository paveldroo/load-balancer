package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	backends          []string
	statuses          *sync.Map
	lastIdx           int
	heartbeatInterval time.Duration
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, "you must specify heartbeat duration in seconds\n")
		os.Exit(1)
	}

	heartbeatInterval, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "you should use number in heartbeat duration, got: %s\n", args[0])
		os.Exit(1)
	}

	cfg := initConfig(heartbeatInterval)

	heartbeat(cfg)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(os.Stdout, "Received request from %s\n", strings.Split(r.RemoteAddr, ":")[0])
		fmt.Fprintf(os.Stdout, "%s %s %s\n", r.Method, r.URL, r.Proto)
		fmt.Fprintf(os.Stdout, "Host: %s\n", r.Host)
		fmt.Fprintf(os.Stdout, "User-Agent: %s\n", r.UserAgent())
		fmt.Fprintf(os.Stdout, "Accept: %s\n", r.Header.Get("Accept"))

		resp, err := http.Get(backend(cfg))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error making request to backend: %s\n", err)
		}

		fmt.Fprintf(os.Stdout, "\nResponse from server: %s %s\n", resp.Proto, resp.Status)

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading response body: %s\n", err.Error())
		}
		fmt.Fprintf(os.Stdout, "\n%s\n", body)

		b := string(body)
		fmt.Fprint(w, b)
	})

	log.Fatal(http.ListenAndServe(":80", nil))
}

func initConfig(heartbeatInterval int) *Config {
	backends := []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"}
	statuses := sync.Map{}
	for i := range backends {
		statuses.Store(i, true)
	}

	return &Config{
		backends:          backends,
		statuses:          &statuses,
		lastIdx:           0,
		heartbeatInterval: time.Duration(heartbeatInterval) * time.Second,
	}
}

func backend(cfg *Config) string {
	for {
		cfg.lastIdx++

		if cfg.lastIdx > len(cfg.backends)-1 {
			cfg.lastIdx = 0
		}

		isLive, _ := cfg.statuses.Load(cfg.lastIdx)
		if isLive.(bool) {
			break
		}

		fmt.Fprintf(os.Stdout, "%s is not available, searching for next live host...\n", cfg.backends[cfg.lastIdx])
		time.Sleep(1 * time.Second)
	}

	return cfg.backends[cfg.lastIdx]
}

func heartbeat(cfg *Config) {
	for i, host := range cfg.backends {
		go func(hostIdx int, statuses *sync.Map, host string) {
			t := time.Tick(cfg.heartbeatInterval)

			for range t {
				resp, err := http.Get(host)
				if err != nil {
					fmt.Fprintf(os.Stderr, "requesting host %s, error: %s\n", host, err)
					statuses.Store(hostIdx, false)

					continue
				}

				statuses.Store(hostIdx, resp.StatusCode != http.StatusOK)

				statuses.Store(hostIdx, true)
				resp.Body.Close()
			}
		}(i, cfg.statuses, host)
	}
}
