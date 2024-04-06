package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {

	fname := flag.String("filename", "./assets/file.mp3", "path of the audio file")
	flag.Parse()
	file, err := os.Open(*fname)
	if err != nil {

		log.Fatal(err)

	}

	ctn, err := io.ReadAll(file)
	if err != nil {

		log.Fatal(err)

	}

	connPool := NewConnectionPool()

	go Stream(connPool, ctn)

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
