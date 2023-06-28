package abstraction

import (
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
)

type Ec struct {
	state         *procstate.ProcState
	queue         *queue.Queue
	abstractionId string
	abstractions  *map[string]Abstraction
	trusted       *pb.ProcessId
	lastts        int
	ts            int
}

func NewEc(state *procstate.ProcState, queue *queue.Queue, abstractionId string, abstractions *map[string]Abstraction) *Ec {

	ec := &Ec{
		state:         state,
		queue:         queue,
		abstractionId: abstractionId,
		abstractions:  abstractions,
		trusted:       state.GetHighestRankingProcess(),
		ts:            int(state.CurrentProcId.Rank),
	}

	pl := NewPl(ec.state, ec.queue, ec.abstractionId+".pl", ec.abstractions)
	beb := NewBeb(ec.state, ec.queue, ec.abstractionId+".beb")
	eld := NewEld(ec.state, ec.queue, ec.abstractionId+".eld", ec.abstractions)
	RegisterAbstraction(ec.abstractions, pl.abstractionId, pl)
	RegisterAbstraction(ec.abstractions, beb.abstractionId, beb)
	RegisterAbstraction(ec.abstractions, eld.abstractionId, eld)

	return ec
}

func (ec *Ec) Handle(msg *pb.Message) error {

	if msg == nil {
		return fmt.Errorf("%s handler received nil message", ec.abstractionId)
	}

	log.Printf("[%s got message]: {%+v}\n\n", ec.abstractionId, msg)

	// Only need to perform the check of one of those two change.
	switch msg.GetType() {
	case pb.Message_ELD_TRUST:
		ec.trusted = msg.GetEldTrust().GetProcess()
		if ec.trusted == ec.state.CurrentProcId {
			ec.ts += len(ec.state.Processes)
			beb := pb.Message{
				Type:              pb.Message_BEB_BROADCAST,
				FromAbstractionId: ec.abstractionId,
				ToAbstractionId:   Next(ec.abstractionId, "beb"),
				BebBroadcast: &pb.BebBroadcast{
					Message: &pb.Message{
						Type:              pb.Message_EC_INTERNAL_NEW_EPOCH,
						FromAbstractionId: ec.abstractionId,
						ToAbstractionId:   ec.abstractionId,
						EcInternalNewEpoch: &pb.EcInternalNewEpoch{
							Timestamp: int32(ec.ts),
						},
					},
				},
			}
			trigger(ec.state, ec.queue, &beb)
		}

	case pb.Message_BEB_DELIVER:

		if msg.GetBebDeliver().GetMessage().GetType() != pb.Message_EC_INTERNAL_NEW_EPOCH {
			break
		}
		if (msg.GetBebBroadcast().GetMessage().GetEcInternalNewEpoch().GetTimestamp() > int32(ec.lastts)) &&
			(msg.GetBebDeliver().GetSender() == ec.trusted) {
			ec.lastts = int(msg.GetBebBroadcast().GetMessage().GetEcInternalNewEpoch().GetTimestamp())
			startEpoch := pb.Message{
				Type:              pb.Message_EC_START_EPOCH,
				FromAbstractionId: ec.abstractionId,
				ToAbstractionId:   Previous(ec.abstractionId),
				EcStartEpoch: &pb.EcStartEpoch{
					NewTimestamp: int32(ec.lastts),
					NewLeader:    msg.GetBebDeliver().GetSender(),
				},
			}
			trigger(ec.state, ec.queue, &startEpoch)

		} else {
			plSend := pb.Message{
				Type:              pb.Message_PL_SEND,
				FromAbstractionId: ec.abstractionId,
				ToAbstractionId:   Next(ec.abstractionId, "pl"),
				PlSend: &pb.PlSend{
					Destination: msg.GetBebDeliver().GetSender(),
					Message: &pb.Message{
						Type:              pb.Message_EC_INTERNAL_NACK,
						FromAbstractionId: ec.abstractionId,
						ToAbstractionId:   ec.abstractionId,
						EcInternalNack:    &pb.EcInternalNack{},
					},
				},
			}
			trigger(ec.state, ec.queue, &plSend)
		}

	case pb.Message_PL_DELIVER:
		if msg.GetBebDeliver().GetMessage().GetType() != pb.Message_EC_INTERNAL_NACK {
			break
		}
		if ec.state.CurrentProcId == ec.trusted {
			ec.ts += len(ec.state.Processes)
			beb := pb.Message{
				Type:              pb.Message_BEB_BROADCAST,
				FromAbstractionId: ec.abstractionId,
				ToAbstractionId:   Next(ec.abstractionId, "beb"),
				BebBroadcast: &pb.BebBroadcast{
					Message: &pb.Message{
						Type:              pb.Message_EC_INTERNAL_NEW_EPOCH,
						FromAbstractionId: ec.abstractionId,
						ToAbstractionId:   ec.abstractionId,
						EcInternalNewEpoch: &pb.EcInternalNewEpoch{
							Timestamp: int32(ec.ts),
						},
					},
				},
			}
			trigger(ec.state, ec.queue, &beb)
		}
	}
	return nil
}
