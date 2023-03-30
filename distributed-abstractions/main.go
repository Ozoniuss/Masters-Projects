package main

import (
	"fmt"
	"net"

	"hw/abstraction"
	abs "hw/abstraction"
	log "hw/log"
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
	var port, index int
	initFlags(&port, &index)

	var host = HOST

	var logg = log.InitLog(log.DEBUG)
	state := state.NewProcState(&logg)

	// Deal with handshake separately because it's easier.
	err := handshake(state, &logg, host, port, index)
	if err != nil {
		logg.Errorf("error during handshake: %s", err.Error())
		return
	}

	logg.Infof("System initialized. Process internal state: %+v", state)

	/*
		These only occur once the handshake is complete.
	*/

	queue := queue.NewQueue(1000)

	abstractions := abs.InitAbstractions(state)

	// Register perfect link.
	pl := abstraction.NewPl(state, queue, &logg)
	abstraction.RegisterAbstraction(abstractions, abs.APP_PL, pl)

	app := abstraction.NewApp(state, queue, logg)
	abstraction.RegisterAbstraction(abstractions, abs.APP, app)

	appbeb := abstraction.NewAppBeb(state, queue, logg)
	abstraction.RegisterAbstraction(abstractions, abs.APP_BEB, appbeb)

	qprocessor := qprocessor.NewQueueProcessor(abstractions, logg)
	qprocessor.Start(queue)

	lis, err := net.Listen(TCP, fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	err = pl.ReadSocket(lis)
	if err != nil {
		logg.Errorf("got error listening to socket: %s", err.Error())
	}
}
