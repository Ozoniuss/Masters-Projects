package abstraction

import (
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
	"strconv"
	"strings"
)

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

	fmt.Println("GENERATED CONSENSUS")

	l := state.GetHighestRankingProcess()
	uc := &Uc{
		state:         state,
		queue:         queue,
		abstractionId: abstractionId,
		abstractions:  abstractions,
		val:           &pb.Value{Defined: false},
		ets:           0,
		leader:        l,
		newts:         0,
		newleader:     nil,
		proposed:      false,
		decided:       false,
	}

	fmt.Println("ABSTRACTIONS,", uc.abstractions)

	ep := NewEp(state, queue, fmt.Sprintf("%s.ep[%d]", abstractionId, 0), abstractions,
		EpState{valts: 0, val: &pb.Value{Defined: false}}, l, 0)
	fmt.Println("new EP")
	ec := NewEc(state, queue, abstractionId+".ec", ep.abstractions)
	fmt.Println("new EC")
	RegisterAbstraction(uc.abstractions, ep.abstractionId, ep)
	fmt.Println("2")
	RegisterAbstraction(uc.abstractions, ec.abstractionId, ec)
	fmt.Println("RETURNED CONSENSUS")
	return uc
}

func (uc *Uc) Handle(msg *pb.Message) error {

	if msg == nil {
		return fmt.Errorf("%s handler received nil message", uc.abstractionId)
	}

	log.Printf("[%s got message]: {%+v}\n\n", uc.abstractionId, msg)

	switch msg.GetType() {
	case pb.Message_UC_PROPOSE:
		uc.val = msg.GetUcPropose().GetValue()
		uc.check()
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
					Value: uc.val,
				},
			}
			trigger(uc.state, uc.queue, &decide)
		}
	}
	return nil
}

func (uc *Uc) check() {
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
