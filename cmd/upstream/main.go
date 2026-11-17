package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello from upstream")
}
func echoHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	w.Write(body)
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		fmt.Fprintln(w, "slow response from upstream")
	case <-r.Context().Done():
		log.Printf("slow request cancelled: %v", r.Context().Err())
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/echo", echoHandler)
	mux.HandleFunc("/slow", slowHandler)

	log.Println("upstream starting on :9090")
	if err := http.ListenAndServe(":9090", mux); err != nil {
		log.Fatal("error staring server for upstream")
	}

}
