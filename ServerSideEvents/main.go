package main

// source: https://www.youtube.com/watch?v=nvijc5J-JAQ

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"time"
)

//go:embed index.html
var indexHTML []byte

func main() {
	http.HandleFunc("/events", sseHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(indexHTML)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("unable to start server %s", err.Error())
	}

}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	w.Header().Set("Access-Control-Allow-Origin", "*")

	rc := http.NewResponseController(w)

	clientGone := r.Context().Done()

	t := time.NewTicker(time.Second)
	var count int = 0
	defer t.Stop()

	for {
		select {
		case <-clientGone:
			fmt.Println("client has disconnected")
			return
		case <-t.C:
			if _, err := fmt.Fprintf(w, "event:test\ndata:Streaming Event %d\n\n", count); err != nil {
				log.Printf("error in writing")
			}
			rc.Flush()
			count++
		}
	}

}
