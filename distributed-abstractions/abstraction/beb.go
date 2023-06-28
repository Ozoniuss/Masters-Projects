package abstraction

import (
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
)

type Beb struct {
	state         *procstate.ProcState
	queue         *queue.Queue
	abstractionId string
}

func NewBeb(state *procstate.ProcState, queue *queue.Queue, abstractionId string) *Beb {
	return &Beb{
		state:         state,
		queue:         queue,
		abstractionId: abstractionId,
	}
}

func (appbeb *Beb) Handle(msg *pb.Message) error {

	if msg == nil {
		return fmt.Errorf("%s handler received nil message", appbeb.abstractionId)
	}

	log.Printf("[%s got message]: {%+v}\n\n", appbeb.abstractionId, msg)

	switch msg.GetType() {

	// When receiving beb broadcast, forward to the beb broadcast perfect link.
	case pb.Message_BEB_BROADCAST:

		// Trigger a perfect link send to all processes except the hub.
		for _, proc := range appbeb.state.Processes {
			if proc.Owner == "hub" {
				continue
			}
			plsend := pb.Message{
				Type:              pb.Message_PL_SEND,
				FromAbstractionId: msg.GetToAbstractionId(),
				ToAbstractionId:   Next(appbeb.abstractionId, "pl"),
				SystemId:          appbeb.state.SystemId,
				MessageUuid:       msg.GetMessageUuid(),
				PlSend: &pb.PlSend{
					Message:     msg.GetBebBroadcast().GetMessage(),
					Destination: proc,
				},
			}
			trigger(appbeb.state, appbeb.queue, &plsend)
		}

	// When receiving a pl deliver, generate a beb deliver message and forward
	//it to app.
	case pb.Message_PL_DELIVER:
		bebdeliver := pb.Message{
			Type:              pb.Message_BEB_DELIVER,
			FromAbstractionId: msg.GetToAbstractionId(),
			ToAbstractionId:   Previous(msg.GetToAbstractionId()),
			SystemId:          appbeb.state.SystemId,
			MessageUuid:       msg.GetMessageUuid(),
			BebDeliver: &pb.BebDeliver{
				Sender:  msg.GetPlDeliver().GetSender(),
				Message: msg.GetPlDeliver().GetMessage(),
			},
		}
		trigger(appbeb.state, appbeb.queue, &bebdeliver)
	}

	return nil
}
