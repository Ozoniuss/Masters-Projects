package state

import (
	"fmt"
	pb "hw/protobuf"

	log "hw/log"
)

type ProcState struct {
	// All processes now about all processes within this abstraction.
	Processes []*pb.ProcessId
	// Closes the process gracefully.
	CurrentProcId *pb.ProcessId
	Quit          chan struct{}
	// The id of the system, as sent by the hub.
	SystemId string
	Logg     *log.Logger
}

func NewProcState(l *log.Logger) *ProcState {
	return &ProcState{
		Logg: l,
		Quit: make(chan struct{}, 1),
	}
}

func (p *ProcState) Name() string {
	return fmt.Sprintf("%s-%d", p.CurrentProcId.Owner, p.CurrentProcId.Index)
}
