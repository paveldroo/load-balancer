package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Config struct {
	backends []string
	lastIdx  int
}

var cfg = Config{
	backends: []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083"},
	lastIdx:  0,
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(os.Stdout, "Received request from %s\n", strings.Split(r.RemoteAddr, ":")[0])
		fmt.Fprintf(os.Stdout, "%s %s %s\n", r.Method, r.URL, r.Proto)
		fmt.Fprintf(os.Stdout, "Host: %s\n", r.Host)
		fmt.Fprintf(os.Stdout, "User-Agent: %s\n", r.UserAgent())
		fmt.Fprintf(os.Stdout, "Accept: %s\n", r.Header.Get("Accept"))

		resp, err := http.Get(backend())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error making request to backend: %s", err)
		}

		fmt.Fprintf(os.Stdout, "\nResponse from server: %s %s\n", resp.Proto, resp.Status)

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading response body: %s", err.Error())
		}
		fmt.Fprintf(os.Stdout, "\n%s\n", body)

		fmt.Fprintf(w, string(body))
	})

	log.Fatal(http.ListenAndServe(":80", nil))
}

func backend() string {
	if cfg.lastIdx > len(cfg.backends)-1 {
		cfg.lastIdx = 0
	}

	b := cfg.backends[cfg.lastIdx]
	cfg.lastIdx++

	return b
}
