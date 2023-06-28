package main

import (
	"fmt"
	"hw/log"
	"net"

	abs "hw/abstraction"
	"hw/qprocessor"
	"hw/queue"
	"hw/state"
)

const (
	HUB_ADDRESS = "127.0.0.1:5000"
	TCP         = "tcp"
	OWNER       = "wtf"
	HOST        = "127.0.0.1"
)

const MAX_MSG_SIZE = 256 * 256 * 256 * 256

func main() {

	// Port and process index are specified via flags.
	var port, index, loglevel int
	initFlags(&port, &index, &loglevel)

	var host = HOST

	state := state.NewProcState(host, int32(port))
	queue := queue.NewQueue(1000)
	abstractions := abs.InitAbstractions(state)

	// Can be used to stop the message
	stopq := make(chan struct{}, 1)

	// Do an initial handshake before reading socker.
	err := handshake(state, host, port, index)
	if err != nil {
		log.Printf("error during handshake: %s", err.Error())
	}

	app := abs.NewApp(state, queue, &abstractions, stopq)
	abs.RegisterAbstraction(&abstractions, abs.APP, app)

	// Get the perfect link.
	pl := abstractions["app.pl"].(*abs.Pl)

	qprocessor := qprocessor.NewQueueProcessor(abstractions, stopq)
	qprocessor.Start(queue)

	lis, err := net.Listen(TCP, fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	err = pl.ReadSocket(lis)
	if err != nil {
		log.Printf("got error listening to socket: %s\n", err.Error())
	}
}
