package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// sleep simulates each thread taking a longer time to process the input.
func sleep(id int) {
	if id == 0 {
		time.Sleep(200 * time.Millisecond)
	}
	if id == 1 {
		time.Sleep(800 * time.Millisecond)
	}
	if id == 2 {
		time.Sleep(2 * time.Second)
	}
}

// processInput processes the received keyboard input from each client.
//
// Note that if the server closes the connection, it is no longer possible
// to read the keyboard input that was sent by the tcp server. So if the
// server sent "a", "b", "c", "d", "e", but the client only managed to process
// up to "b" until the server closed, it's no longer possible to read the
// remaining characters. For this reason, messages will be added to a
// processing queue, and then be processed in the background.
func processInput(id int, conn net.Conn) {

	var mutex = &sync.Mutex{}
	var wg = &sync.WaitGroup{}
	q := []byte{}
	out := ""

	closed := false
	wg.Add(1)

	go func() {
		for {
			if closed && len(q) == 0 {
				break
			}
			if len(q) > 0 {
				mutex.Lock()
				top := q[:1]
				q = q[1:]
				mutex.Unlock()

				sleep(id)
				out += string(top)
				fmt.Printf("\rgot: %s", out)
			}
		}
		wg.Done()
	}()

	for {
		var b = make([]byte, 1)
		_, err := conn.Read(b)
		if err != nil {
			// There is no need to protect this since the other thread only
			// reads it.
			closed = true
			if err != nil {
				fmt.Println("\rclient closed")
				// format nicely
				fmt.Printf("\rgot: %s", out)
				break
			}
		}
		// Add the message to the processing queue. Note that we only want
		// to protect the writes to the queue in this and the worker thread,
		// to have the least impact on performance.
		mutex.Lock()
		q = append(q, b[0])
		mutex.Unlock()
	}

	wg.Wait()
	fmt.Println("\ndone")
}

func main() {

	//establish connection
	conn, err := net.Dial("tcp", "127.0.0.1:9966")
	if err != nil {
		panic(err)
	}

	// The server will communicate the process id, which determines how long
	// the operation takes.
	idbyte := make([]byte, 1)
	_, err = conn.Read(idbyte)
	if err != nil {
		panic(err)
	}

	id := int(idbyte[0])
	fmt.Println(id)

	processInput(int(idbyte[0]), conn)
}
