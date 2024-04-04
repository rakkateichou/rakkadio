package main

import (
  "sync"
  "bytes"
  "time"
  "io"
)

const (
	BUFFERSIZE = 8192

	//formula for delay = track_duration * buffer_size / aac_file_size
	DELAY = 150
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

func Stream(connectionPool *ConnectionPool, content []byte) {

	buffer := make([]byte, BUFFERSIZE)

	for {

		// clear() is a new builtin function introduced in go 1.21. Just reinitialize the buffer if on a lower version.
		clear(buffer)
		tempfile := bytes.NewReader(content)
		ticker := time.NewTicker(time.Millisecond * DELAY)

		for range ticker.C {

			_, err := tempfile.Read(buffer)

			if err == io.EOF {

				ticker.Stop()
				break

			}

			connectionPool.Broadcast(buffer)

		}

	}

}
