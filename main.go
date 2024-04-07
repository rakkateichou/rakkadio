package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {

	connPool := NewConnectionPool()

	songEnded := make(chan struct{})
	songName := make(chan string)
	
	go Stream(connPool, songEnded, songName)
	go TopUpMusic(songEnded, songName)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

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

  log.Println("Listening on http://localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
