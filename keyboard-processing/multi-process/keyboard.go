package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	fmt.Println("start keyboard server")

	s, err := net.Listen("tcp", "127.0.0.1:9966")
	if err != nil {
		// don't panic in production.
		panic(err)
	}
	defer s.Close()

	// Client connections. We are expecting 3 client connections.
	var conns [3]net.Conn
	connected := 0
	for connected < 3 {
		client, err := s.Accept()
		if err != nil {
			panic(err)
		}
		// Send the id to the client.
		client.Write([]byte{byte(connected)})
		fmt.Printf("client %d connected: address %s\n", connected, client.RemoteAddr())
		conns[connected] = client
		connected++
	}

	// Read keyboard input until end of line.
	r := bufio.NewReader(os.Stdin)
	for {
		chrs, _, err := r.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				panic(err)
			}
		}
		// To each client, send the processed character.
		for _, chr := range chrs {
			conns[0].Write([]byte{chr})
			conns[1].Write([]byte{chr})
			conns[2].Write([]byte{chr})
		}
	}
}
