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

	Host string
	Port int32
}

func NewProcState(host string, port int32) *ProcState {
	return &ProcState{
		Quit: make(chan struct{}, 1),
		Host: host,
		Port: port,
	}
}

func (p *ProcState) Name() string {
	if p.CurrentProcId == nil {
		return ""
	}
	return fmt.Sprintf("%s-%d", p.CurrentProcId.Owner, p.CurrentProcId.Index)
}

func (p *ProcState) GetProcessesAsMap() map[*pb.ProcessId]struct{} {
	all := make(map[*pb.ProcessId]struct{}, len(p.Processes))
	for _, p := range p.Processes {
		all[p] = struct{}{}
	}
	return all
}

func (p *ProcState) GetHighestRankingProcess() *pb.ProcessId {
	var ret *pb.ProcessId
	var maxRank = -1

	for _, p := range p.Processes {
		if p.Rank > int32(maxRank) {
			ret = p
			maxRank = int(p.Rank)
		}
	}
	return ret
}
