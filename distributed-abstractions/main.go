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

	state := state.NewProcState()

	// Deal with handshake separately because it's easier.
	err := handshake(state, host, port, index)
	if err != nil {
		log.Printf("error during handshake: %s\n\n", err.Error())
		return
	}

	log.Printf("System initialized. Process internal state: %+v\n\n", state)

	/*
		These only occur once the handshake is complete.
	*/

	queue := queue.NewQueue(1000)

	abstractions := abs.InitAbstractions(state)

	pl := abs.NewPl(state, queue, &abstractions)
	abs.RegisterAbstraction(&abstractions, abs.APP_PL, pl)

	bebpl := abs.NewPl(state, queue, &abstractions)
	abs.RegisterAbstraction(&abstractions, abs.APP_BEB_PL, bebpl)

	app := abs.NewApp(state, queue, &abstractions)
	abs.RegisterAbstraction(&abstractions, abs.APP, app)

	appbeb := abs.NewAppBeb(state, queue, "app.beb")
	abs.RegisterAbstraction(&abstractions, abs.APP_BEB, appbeb)

	qprocessor := qprocessor.NewQueueProcessor(abstractions)
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
