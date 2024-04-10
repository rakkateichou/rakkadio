package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {

	connPool := NewConnectionPool()

	songEnded := make(chan struct{})
	songPath := make(chan string)

	go Stream(connPool, songEnded, songPath)
	go TopUpMusic(songEnded, songPath)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World :)"))
	})

	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Content-Type", "audio/mpeg")
		w.Header().Add("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {

			log.Println("Could not create flusher")

		}

		connection := &Connection{bufferChannel: make(chan []byte), buffer: make([]byte, BUFFERSIZE)}
		connPool.AddConnection(connection)
		log.Printf("%s has connected to the audio stream\n", r.Host)

		for {

			buf := <-connection.bufferChannel
			if _, err := w.Write(buf); err != nil {

				connPool.DeleteConnection(connection)
				log.Printf("%s's connection to the audio stream has been closed\n", r.Host)
				return

			}
			flusher.Flush()
			clear(connection.buffer)

		}
	})

	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		
		w.Header().Add("Content-Type", "application/json")

		trackInfo := GetCurrentTrackInfo()

		res, err := json.Marshal(trackInfo)
		if err != nil {
			log.Fatal("Couldn't encode track info")
		}

		w.Write(res)
	})

	log.Println("Listening on http://localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
