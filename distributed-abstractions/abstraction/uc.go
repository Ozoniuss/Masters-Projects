package abstraction

import (
	"fmt"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
	"log"
	"strconv"
	"strings"
)

// Fail-noisy model.

type Uc struct {
	state         *procstate.ProcState
	queue         *queue.Queue
	abstractionId string
	abstractions  *map[string]Abstraction
	ets           int
	newts         int
	leader        *pb.ProcessId
	newleader     *pb.ProcessId
	val           *pb.Value
	proposed      bool
	decided       bool
}

func NewUc(state *procstate.ProcState, queue *queue.Queue, abstractions *map[string]Abstraction, abstractionId string) *Uc {

	l := state.GetHighestRankingProcess()
	uc := &Uc{
		state:         state,
		queue:         queue,
		abstractionId: abstractionId,
		abstractions:  abstractions,

		val: &pb.Value{Defined: false},
		// Start with an epoch consensus timestamp of 0 and leader with highest
		// rank.
		ets:       0,
		leader:    l,
		newts:     0,
		newleader: nil,
		proposed:  false,
		decided:   false,
	}

	ep := NewEp(state, queue, fmt.Sprintf("%s.ep[%d]", abstractionId, uc.ets), uc.abstractions,
		EpState{valts: 0, val: &pb.Value{Defined: false}}, uc.leader, 0)
	ec := NewEc(state, queue, abstractionId+".ec", ep.abstractions)
	RegisterAbstraction(uc.abstractions, ep.abstractionId, ep)
	RegisterAbstraction(uc.abstractions, ec.abstractionId, ec)
	log.Printf("HERE IS THE LEADER: %+v\n\n", l)
	return uc
}

func (uc *Uc) Handle(msg *pb.Message) error {

	switch msg.GetType() {

	case pb.Message_UC_PROPOSE:
		uc.val = msg.GetUcPropose().GetValue()
		uc.check()

		// If I start a new epoch, abort the current epoch.
	case pb.Message_EC_START_EPOCH:
		uc.newts = int(msg.GetEcStartEpoch().GetNewTimestamp())
		uc.newleader = msg.GetEcStartEpoch().GetNewLeader()
		abort := pb.Message{
			Type:              pb.Message_EP_ABORT,
			FromAbstractionId: uc.abstractionId,
			ToAbstractionId:   fmt.Sprintf("%s.ep[%d]", uc.abstractionId, uc.ets),
			EpAbort:           &pb.EpAbort{},
		}
		trigger(uc.state, uc.queue, &abort)
		// uc.check()

		// If the epoch with the timestamp I have was aborted.
	case pb.Message_EP_ABORTED:
		// TODO: Should this condition be written differently?
		if uc.ets != int(msg.GetEpAborted().GetEts()) {
			break
		}
		// if uc.ets != findEpochFromEp(msg.GetFromAbstractionId()) {
		// 	break
		// }
		uc.ets = uc.newts
		uc.leader = uc.newleader
		uc.proposed = false

		// Get the state from the aborted message. Use it as new
		newState := EpState{
			valts: int(msg.GetEpAborted().GetValueTimestamp()),
			val:   msg.GetEpAborted().GetValue(),
		}

		ep := NewEp(uc.state, uc.queue, fmt.Sprintf("%s.ep[%d]", uc.abstractionId, uc.ets),
			uc.abstractions, newState, uc.leader, uc.ets)
		RegisterAbstraction(uc.abstractions, ep.abstractionId, ep)
		uc.check()

	case pb.Message_EP_DECIDE:
		// TODO: Should this condition be written differently?
		if uc.ets != int(msg.GetEpDecide().GetEts()) {
			break
		}
		// if uc.ets != findEpochFromEp(msg.GetFromAbstractionId()) {
		// 	break
		// }
		if !uc.decided {
			uc.decided = true
			decide := pb.Message{
				Type:              pb.Message_UC_DECIDE,
				FromAbstractionId: uc.abstractionId,
				ToAbstractionId:   Previous(uc.abstractionId),
				UcDecide: &pb.UcDecide{
					Value: msg.GetEpDecide().GetValue(),
				},
			}
			trigger(uc.state, uc.queue, &decide)
		}
	}
	return nil
}

func (uc *Uc) check() {
	// If I'm a leader and a have a falue I propose.
	if (uc.leader == uc.state.CurrentProcId) &&
		(!uc.proposed) && (uc.val.GetDefined()) {
		uc.proposed = true
		epPropose := pb.Message{
			Type:              pb.Message_EP_PROPOSE,
			FromAbstractionId: uc.abstractionId,
			ToAbstractionId:   fmt.Sprintf("%s.ep[%d]", uc.abstractionId, uc.ets),
			EpPropose: &pb.EpPropose{
				Value: uc.val,
			},
		}
		log.Printf("PROPOISED.... \n\n")
		trigger(uc.state, uc.queue, &epPropose)
	}
}

func findEpochFromEp(epId string) int {
	parts := strings.Split(epId, ".")
	id := parts[2][3 : len(parts[2])-1]
	idint, err := strconv.Atoi(id)
	if err != nil {
		panic("GETTING EP ID")
	}
	return idint
}
