package abstraction

import (
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
)

type EpState struct {
	valts int
	val   *pb.Value
}

func highest(states map[*pb.ProcessId]EpState) EpState {
	highest := EpState{
		valts: -2,
	}
	for _, state := range states {
		if state.valts > highest.valts {
			highest = state
		}
	}
	return highest
}

// Epoch consensus is only an attempt to reach consensus. It may not be
// terminated and aborted when it doesn't decide  or the next epoch should
// be started.
//
// Only the leader proposes a value and ep decides only when the leader is
// correct.
//
// The higher level algorithm executes multiple ep algorithms until one decides.
//
// Every instance of EP is associated with a timestamp "ts" and a leader "l".
// The leader "l" proposes a value "v". The other processes are not required to
// propose anything.
// Epoch consensus must etrminate when Abort is triggered, which when completed
// the epoch triggers an <Aborted, state> event to the higher-level algorithm.
// The state is used to initialize the next epoch.
//
// Processes initialize an EP at most one at a time, only after the previous
// one aborted or ep-decided. It should have a higher timestamp than all
// instances it initalized previously, with the most recent state.

type Ep struct {
	state         *procstate.ProcState
	queue         *queue.Queue
	abstractionId string
	abstractions  *map[string]Abstraction
	epStates      map[*pb.ProcessId]EpState
	tmpval        *pb.Value
	// State contains the locally stored value and the timestamp during which
	// it was written.
	epState  EpState
	accepted int
	// Timestamp and (potentially unnecesarry) leader of this ep
	ets    int
	leader *pb.ProcessId
	halt   bool
}

func NewEp(state *procstate.ProcState, queue *queue.Queue, abstractionId string, abstractions *map[string]Abstraction, epState EpState, leader *pb.ProcessId, ts int) *Ep {

	ep := &Ep{
		state:         state,
		queue:         queue,
		abstractionId: abstractionId,
		abstractions:  abstractions,
		epState:       epState,
		accepted:      0,
		epStates:      make(map[*pb.ProcessId]EpState),
		tmpval:        &pb.Value{Defined: false},
		halt:          false,
		ets:           ts,
		leader:        leader,
	}

	bebpl := NewPl(ep.state, ep.queue, ep.abstractionId+".beb.pl", ep.abstractions)
	pl := NewPl(ep.state, ep.queue, ep.abstractionId+".pl", ep.abstractions)
	beb := NewBeb(ep.state, ep.queue, ep.abstractionId+".beb")
	RegisterAbstraction(ep.abstractions, pl.abstractionId, pl)
	RegisterAbstraction(ep.abstractions, beb.abstractionId, beb)
	RegisterAbstraction(ep.abstractions, bebpl.abstractionId, bebpl)

	return ep
}

func (ep *Ep) Handle(msg *pb.Message) error {

	if ep.halt {
		log.Printf("%s halting...\n\n", ep.abstractionId)
		return nil
	}

	switch msg.GetType() {
	case pb.Message_EP_PROPOSE:
		ep.tmpval = msg.GetEpPropose().GetValue()
		beb := pb.Message{
			Type:              pb.Message_BEB_BROADCAST,
			FromAbstractionId: ep.abstractionId,
			ToAbstractionId:   Next(ep.abstractionId, "beb"),
			BebBroadcast: &pb.BebBroadcast{
				Message: &pb.Message{
					Type:              pb.Message_EP_INTERNAL_READ,
					FromAbstractionId: ep.abstractionId,
					ToAbstractionId:   ep.abstractionId,
					EpInternalRead:    &pb.EpInternalRead{},
				},
			},
		}
		trigger(ep.state, ep.queue, &beb)

	case pb.Message_BEB_DELIVER:
		switch msg.GetBebDeliver().GetMessage().GetType() {
		// After proposing, the leader tries to read state from all processes.
		case pb.Message_EP_INTERNAL_READ:
			// TODO: is this necessary?
			// if bebdeliver.GetBebDeliver().GetSender() != ep.leader {
			// 	break
			// }
			plSend := pb.Message{
				Type:              pb.Message_PL_SEND,
				FromAbstractionId: ep.abstractionId,
				ToAbstractionId:   Next(ep.abstractionId, "pl"),
				// send the state back to the leader.
				PlSend: &pb.PlSend{
					Destination: msg.GetBebDeliver().GetSender(), // fixed
					Message: &pb.Message{
						Type:              pb.Message_EP_INTERNAL_STATE,
						ToAbstractionId:   ep.abstractionId,
						FromAbstractionId: ep.abstractionId,
						EpInternalState: &pb.EpInternalState{
							ValueTimestamp: int32(ep.epState.valts),
							Value:          ep.epState.val,
						},
					},
				},
			}
			trigger(ep.state, ep.queue, &plSend)

			// Here processes will try to accept the write of the highest
			// state by the leader. They will use the timestamp of the current
			// epoch consensus timestamp/
		case pb.Message_EP_INTERNAL_WRITE:
			// if bebdeliver.GetBebDeliver().GetSender() != ep.leader {
			// 	break
			// }
			ep.epState = EpState{
				valts: ep.ets,
				val:   msg.GetBebDeliver().GetMessage().GetEpInternalWrite().GetValue(),
			}
			// Process accepted the value. Confirm it to the leader with an
			// accept.
			plSend := pb.Message{
				Type:              pb.Message_PL_SEND,
				FromAbstractionId: ep.abstractionId,
				ToAbstractionId:   Next(ep.abstractionId, "pl"),
				PlSend: &pb.PlSend{
					Destination: msg.GetBebDeliver().GetSender(),
					Message: &pb.Message{
						Type:              pb.Message_EP_INTERNAL_ACCEPT,
						ToAbstractionId:   ep.abstractionId,
						FromAbstractionId: ep.abstractionId,
						EpInternalAccept:  &pb.EpInternalAccept{},
					},
				},
			}
			trigger(ep.state, ep.queue, &plSend)

		case pb.Message_EP_INTERNAL_DECIDED:
			// if bebdeliver.GetBebDeliver().GetSender() != ep.leader {
			// 	break
			// }
			decide := pb.Message{
				Type:              pb.Message_EP_DECIDE,
				FromAbstractionId: ep.abstractionId,
				ToAbstractionId:   Previous(ep.abstractionId),
				EpDecide: &pb.EpDecide{
					Ets:   int32(ep.ets),
					Value: msg.GetBebDeliver().GetMessage().GetEpInternalDecided().GetValue(),
				},
			}
			trigger(ep.state, ep.queue, &decide)
		}

	case pb.Message_PL_DELIVER:
		switch msg.GetPlDeliver().GetMessage().GetType() {
		// Count the accepts.
		case pb.Message_EP_INTERNAL_ACCEPT:
			ep.accepted++

			// Technically you would have to queue this because maybe other
			// events are being processed... hope it's ok...
			// TODO: mention this

			if ep.accepted > len(ep.state.Processes)/2 {
				ep.accepted = 0

				// Leader decides on the message when a quorum accepts it.
				beb := pb.Message{
					Type:              pb.Message_BEB_BROADCAST,
					FromAbstractionId: ep.abstractionId,
					ToAbstractionId:   Next(ep.abstractionId, "beb"),
					BebBroadcast: &pb.BebBroadcast{
						Message: &pb.Message{
							Type:              pb.Message_EP_INTERNAL_DECIDED,
							ToAbstractionId:   ep.abstractionId,
							FromAbstractionId: ep.abstractionId,
							EpInternalDecided: &pb.EpInternalDecided{
								Value: ep.tmpval,
							},
						},
					},
				}
				trigger(ep.state, ep.queue, &beb)
			}

		case pb.Message_EP_INTERNAL_STATE:
			ep.epStates[msg.GetPlDeliver().GetSender()] = EpState{
				valts: int(msg.GetPlDeliver().GetMessage().GetEpInternalState().GetValueTimestamp()),
				val:   msg.GetPlDeliver().GetMessage().GetEpInternalState().GetValue(),
			}

			// If I have more than a majority of states, choose the one
			// with the highest timestamp (which implicitly has a given value)
			// for all processes.
			if len(ep.epStates) > len(ep.state.Processes)/2 {

				// Similar to accepted...
				highestState := highest(ep.epStates)
				if highestState.val.GetDefined() {
					ep.tmpval = highestState.val
				}
				ep.epStates = make(map[*pb.ProcessId]EpState)
				// The leader broadcasts that it's written the highest state.
				beb := pb.Message{
					Type:              pb.Message_BEB_BROADCAST,
					FromAbstractionId: ep.abstractionId,
					ToAbstractionId:   Next(ep.abstractionId, "beb"),
					BebBroadcast: &pb.BebBroadcast{
						Message: &pb.Message{
							Type:              pb.Message_EP_INTERNAL_WRITE,
							FromAbstractionId: ep.abstractionId,
							ToAbstractionId:   ep.abstractionId,
							EpInternalWrite: &pb.EpInternalWrite{
								Value: ep.tmpval,
							},
						},
					},
				}
				trigger(ep.state, ep.queue, &beb)
			}
		}

		// Higher level layer may abort for any reason. Trigger an aborted
		// to the higher level and stop all execution when this happens
		// so that states are kept intact.
	case pb.Message_EP_ABORT:
		aborted := pb.Message{
			Type:              pb.Message_EP_ABORTED,
			FromAbstractionId: ep.abstractionId,
			ToAbstractionId:   Previous(ep.abstractionId),
			EpAborted: &pb.EpAborted{
				Ets:            int32(ep.ets),
				ValueTimestamp: int32(ep.epState.valts),
				Value:          ep.epState.val,
			},
		}
		trigger(ep.state, ep.queue, &aborted)
		ep.halt = true
	}

	return nil
}
