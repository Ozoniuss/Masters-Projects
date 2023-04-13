package abstraction

import (
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
)

type AppBeb struct {
	state *procstate.ProcState
	queue *queue.Queue
}

func NewAppBeb(state *procstate.ProcState, queue *queue.Queue) *AppBeb {
	return &AppBeb{
		state: state,
		queue: queue,
	}
}

func (appbeb *AppBeb) Handle(msg *pb.Message) error {

	if msg == nil {
		return fmt.Errorf("%s handler received nil message", APP_BEB)
	}

	log.Printf("%s got message: %+v\n\n", APP_BEB, msg)

	switch msg.GetType() {

	// When receiving beb broadcast, forward to the beb broadcast perfect link.
	case pb.Message_BEB_BROADCAST:

		// Trigger a perfect link send to all processes.
		for _, proc := range appbeb.state.Processes {
			plsend := pb.Message{
				Type:              pb.Message_PL_SEND,
				FromAbstractionId: msg.GetToAbstractionId(),
				ToAbstractionId:   APP_BEB_PL,
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
			ToAbstractionId:   APP,
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
