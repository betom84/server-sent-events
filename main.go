package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Listening at http://localhost:8081")

	http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			fmt.Printf("streaming not supported")
			return
		}

		c := make(chan string)
		defer close(c)

		t := time.NewTicker(1 * time.Second)
		defer t.Stop()

		go func() {
			for {
				select {
				case <-t.C:
					c <- time.Now().Format(time.RFC3339)
					fmt.Print(".")
				case <-r.Context().Done():
					fmt.Println("client connection closed")
					return
				}
			}
		}()

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		for m := range c {
			fmt.Fprintf(w, "event: message\n")
			fmt.Fprintf(w, "data: %s\n", m)
			fmt.Fprint(w, "\n")

			fmt.Print("_")
			flusher.Flush()
		}
	})

	http.Handle("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe(":8081", nil)
}
