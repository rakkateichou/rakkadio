package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/tcolgate/mp3"
)

const (
	BUFFERSIZE = 8192
)

type Connection struct {
	bufferChannel chan []byte
	buffer        []byte
}

type ConnectionPool struct {
	ConnectionMap map[*Connection]struct{}
	mu            sync.Mutex
}

func (cp *ConnectionPool) AddConnection(connection *Connection) {

	defer cp.mu.Unlock()
	cp.mu.Lock()

	cp.ConnectionMap[connection] = struct{}{}

}

func (cp *ConnectionPool) DeleteConnection(connection *Connection) {

	defer cp.mu.Unlock()
	cp.mu.Lock()

	delete(cp.ConnectionMap, connection)

}

func (cp *ConnectionPool) Broadcast(buffer []byte) {

	defer cp.mu.Unlock()
	cp.mu.Lock()

	for connection := range cp.ConnectionMap {

		copy(connection.buffer, buffer)

		select {

		case connection.bufferChannel <- connection.buffer:

		default:

		}

	}

}

func NewConnectionPool() *ConnectionPool {

	connectionMap := make(map[*Connection]struct{})
	return &ConnectionPool{ConnectionMap: connectionMap}

}

func Stream(connectionPool *ConnectionPool, songEnded chan struct{}, nextSong chan string) {

	buffer := make([]byte, BUFFERSIZE)

	for {

		songEnded <- struct{}{}

		// clear() is a new builtin function introduced in go 1.21. Just reinitialize the buffer if on a lower version.
		clear(buffer)
		nextSongPath := <-nextSong
		songfile, err := os.Open(nextSongPath)
		if err != nil {
			log.Fatal("Couldn't open next song file")
		}

		var songBuffer bytes.Buffer
		tee := io.TeeReader(songfile, &songBuffer)

		duration := 0.0
		d := mp3.NewDecoder(tee)
    var f mp3.Frame
    skipped := 0

    for {
        if err := d.Decode(&f, &skipped); err != nil {
            if err == io.EOF {
                break
            }
            return
        }
        duration = duration + f.Duration().Seconds()
    }

		filestat, err := songfile.Stat()
		if err != nil {
			log.Fatal("Couldn't read file stat")
		}

		filesize := filestat.Size()

		//formula for delay in seconds = track_duration * buffer_size / file_size
		ticker := time.NewTicker(time.Millisecond * time.Duration((duration * BUFFERSIZE / float64(filesize)) * 1000))

		songfile.Close()

		log.Println(fmt.Sprintf("Playing %s", nextSongPath))

		for range ticker.C {

			_, err := songBuffer.Read(buffer)

			if err == io.EOF {
				log.Println(fmt.Sprintf("%s ended", nextSongPath))
				ticker.Stop()
				break
			}

			connectionPool.Broadcast(buffer)

		}

	}

}
