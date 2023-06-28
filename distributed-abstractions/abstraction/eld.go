package abstraction

import (
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
)

type Eld struct {
	state         *procstate.ProcState
	queue         *queue.Queue
	abstractionId string
	abstractions  *map[string]Abstraction
	suspected     map[*pb.ProcessId]struct{}
	leader        *pb.ProcessId
}

func NewEld(state *procstate.ProcState, queue *queue.Queue, abstractionId string, abstractions *map[string]Abstraction) *Eld {

	eld := &Eld{
		state:         state,
		queue:         queue,
		abstractionId: abstractionId,
		abstractions:  abstractions,
		suspected:     make(map[*pb.ProcessId]struct{}, len(state.Processes)),
		leader:        nil,
	}

	epfd := NewEpfd(eld.state, eld.queue, eld.abstractionId+".epfd", eld.abstractions)
	RegisterAbstraction(eld.abstractions, epfd.abstractionId, epfd)
	return eld
}

func (eld *Eld) check() {
	var leader *pb.ProcessId
	maxRank := -1
	for _, proc := range eld.state.Processes {
		// choose a leader from the processes that are not suspected
		if _, ok := eld.suspected[proc]; !ok {
			if proc.Rank > int32(maxRank) {
				leader = proc
				maxRank = int(proc.Rank)
			}
		}
	}

	if eld.leader != leader {
		eld.leader = leader
		eldTrust := pb.Message{
			Type:              pb.Message_ELD_TRUST,
			FromAbstractionId: eld.abstractionId,
			ToAbstractionId:   Previous(eld.abstractionId),
			EldTrust: &pb.EldTrust{
				Process: eld.leader,
			},
		}
		trigger(eld.state, eld.queue, &eldTrust)
	}
}

func (eld *Eld) Handle(msg *pb.Message) error {

	if msg == nil {
		return fmt.Errorf("%s handler received nil message", eld.abstractionId)
	}

	log.Printf("[%s got message]: {%+v}\n\n", eld.abstractionId, msg)

	// Only need to perform the check of one of those two change.
	switch msg.GetType() {
	case pb.Message_EPFD_SUSPECT:
		eld.suspected[msg.GetEpfdSuspect().GetProcess()] = struct{}{}
		eld.check()
	case pb.Message_EPFD_RESTORE:
		delete(eld.suspected, msg.GetEpfdSuspect().GetProcess())
		eld.check()
	}
	return nil
}
