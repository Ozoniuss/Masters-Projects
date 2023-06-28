package abstraction

import (
	"fmt"
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
		valts: -1,
	}
	for _, state := range states {
		if state.valts > highest.valts {
			highest = state
		}
	}
	return highest
}

type Ep struct {
	state         *procstate.ProcState
	queue         *queue.Queue
	abstractionId string
	abstractions  *map[string]Abstraction
	epStates      map[*pb.ProcessId]EpState
	tmpval        *pb.Value
	epState       EpState
	accepted      int
	ets           int
	leader        *pb.ProcessId
	halt          bool
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

	pl := NewPl(ep.state, ep.queue, ep.abstractionId+".pl", ep.abstractions)
	beb := NewBeb(ep.state, ep.queue, ep.abstractionId+".beb")
	RegisterAbstraction(ep.abstractions, pl.abstractionId, pl)
	RegisterAbstraction(ep.abstractions, beb.abstractionId, beb)

	return ep
}

func (ep *Ep) Handle(msg *pb.Message) error {

	if msg == nil {
		return fmt.Errorf("%s handler received nil message", ep.abstractionId)
	}

	log.Printf("[%s got message]: {%+v}\n\n", ep.abstractionId, msg)

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
		switch bebdeliver := msg.GetBebDeliver().GetMessage(); bebdeliver.GetType() {
		case pb.Message_EP_INTERNAL_READ:
			// TODO: is this necessary?
			if bebdeliver.GetBebDeliver().GetSender() != ep.leader {
				break
			}
			plSend := pb.Message{
				Type:              pb.Message_PL_SEND,
				FromAbstractionId: ep.abstractionId,
				ToAbstractionId:   Next(ep.abstractionId, "pl"),
				PlSend: &pb.PlSend{
					Destination: bebdeliver.GetBebDeliver().GetSender(),
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

		case pb.Message_EP_INTERNAL_WRITE:
			if bebdeliver.GetBebDeliver().GetSender() != ep.leader {
				break
			}
			ep.epState = EpState{
				valts: ep.ets,
				val:   bebdeliver.GetBebDeliver().GetMessage().GetEpInternalWrite().GetValue(),
			}
			plSend := pb.Message{
				Type:              pb.Message_PL_SEND,
				FromAbstractionId: ep.abstractionId,
				ToAbstractionId:   Next(ep.abstractionId, "pl"),
				PlSend: &pb.PlSend{
					Destination: bebdeliver.GetBebDeliver().GetSender(),
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
			if bebdeliver.GetBebDeliver().GetSender() != ep.leader {
				break
			}
			decide := pb.Message{
				Type:              pb.Message_EP_DECIDE,
				FromAbstractionId: ep.abstractionId,
				ToAbstractionId:   Previous(ep.abstractionId),
				EpDecide: &pb.EpDecide{
					Ets:   int32(ep.ets),
					Value: bebdeliver.GetBebDeliver().GetMessage().GetEpInternalDecided().GetValue(),
				},
			}
			trigger(ep.state, ep.queue, &decide)
		}

	case pb.Message_PL_DELIVER:
		switch msg.GetPlDeliver().GetMessage().GetType() {
		case pb.Message_EP_INTERNAL_ACCEPT:
			ep.accepted++
			if ep.accepted > len(ep.state.Processes)/2 {
				ep.accepted = 0
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
			if len(ep.epStates) > len(ep.state.Processes)/2 {

				// Is ts a typo?
				highestState := highest(ep.epStates)
				// TODO: maybe check againts null?
				if highestState.val.GetDefined() {
					ep.tmpval = highestState.val
				}
				ep.epStates = make(map[*pb.ProcessId]EpState)
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
