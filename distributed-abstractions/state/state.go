package state

import (
	"fmt"
	pb "hw/protobuf"
)

type ProcState struct {
	// All processes now about all processes within this abstraction.
	Processes []*pb.ProcessId
	// Closes the process gracefully.
	CurrentProcId *pb.ProcessId
	Quit          chan struct{}
	// The id of the system, as sent by the hub.
	SystemId string
}

func NewProcState() *ProcState {
	return &ProcState{
		Quit: make(chan struct{}, 1),
	}
}

func (p *ProcState) Name() string {
	return fmt.Sprintf("%s-%d", p.CurrentProcId.Owner, p.CurrentProcId.Index)
}

func (p *ProcState) GetProcessesAsMap() map[*pb.ProcessId]struct{} {
	all := make(map[*pb.ProcessId]struct{}, len(p.Processes))
	for _, p := range p.Processes {
		all[p] = struct{}{}
	}
	return all
}
