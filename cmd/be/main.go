package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(os.Stdout, "Received request from %s\n", strings.Split(r.RemoteAddr, ":")[0])
		fmt.Fprintf(os.Stdout, "%s %s %s\n", r.Method, r.URL, r.Proto)
		fmt.Fprintf(os.Stdout, "Host: %s\n", r.Host)
		fmt.Fprintf(os.Stdout, "User-Agent: %s\n", r.UserAgent())
		fmt.Fprintf(os.Stdout, "Accept: %s\n", r.Header.Get("Accept"))

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Hello From Backend Server\n")

		fmt.Fprintf(os.Stdout, "\nReplied with a hello message\n\n")
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}
