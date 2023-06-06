package abstraction

import (
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
)

type Nnar struct {
	state      *procstate.ProcState
	queue      *queue.Queue
	ts         int
	wr         int
	val        int32
	acks       int
	writeval   int32
	rid        int
	readlist   []int32
	readval    int
	reading    bool
	registerId int
}

// NewNnar also handles of nnar.Init()
func NewNnar(state *procstate.ProcState, queue *queue.Queue, registerId int) *Nnar {
	return &Nnar{
		state:      state,
		queue:      queue,
		ts:         0,
		wr:         0,
		val:        0,
		writeval:   0,
		rid:        0,
		readlist:   make([]int32, len(state.Processes)),
		readval:    0,
		reading:    false,
		registerId: registerId,
	}
}

func (nnar *Nnar) Handle(msg *pb.Message) error {

	if msg == nil {
		return fmt.Errorf("%s handler received nil message", APP_NNAR)
	}

	log.Printf("%s got message: %+v\n\n", APP_NNAR, msg)

	switch msg.GetType() {

	// Handling NNAR_WRITE, received after emitting an APP_WRITE from the hub.
	case pb.Message_NNAR_WRITE:
		nnar.rid++
		nnar.writeval = msg.GetNnarWrite().GetValue().GetV()
		nnar.acks = 0
		nnar.readlist = make([]int32, len(nnar.state.Processes))
		bebBroadcast := pb.Message{
			Type:              pb.Message_BEB_BROADCAST,
			FromAbstractionId: msg.GetToAbstractionId(),
			ToAbstractionId:   Next(msg.GetToAbstractionId(), "beb"),
			BebBroadcast: &pb.BebBroadcast{
				Message: &pb.Message{
					Type:              pb.Message_NNAR_INTERNAL_READ,
					FromAbstractionId: msg.GetToAbstractionId(),
					ToAbstractionId:   msg.GetToAbstractionId(),
					NnarInternalRead: &pb.NnarInternalRead{
						ReadId: int32(nnar.rid),
					},
				},
			},
		}
		trigger(nnar.state, nnar.queue, &bebBroadcast)

	case pb.Message_BEB_DELIVER:
		plSend := pb.Message{
			Type:              pb.Message_PL_SEND,
			FromAbstractionId: msg.GetToAbstractionId(),
			ToAbstractionId:   Next(msg.GetToAbstractionId(), "pl"),
			PlSend: &pb.PlSend{
				Destination: msg.GetBebDeliver().GetSender(),
				// ShOULD SEND INTERNAL VALUE
				Message: msg.GetBebBroadcast().GetMessage(),
			},
		}
		trigger(nnar.state, nnar.queue, &plSend)

		// case pb.Message_PL_DELIVER:
		// 	// msg.GetPlDeliver().GetMessage().GetNnarInternalValue()
		// 	// nnar.readlist[]
		// 	msg.GetPlDeliver().GetMessage().GetNnarRead().

		// case pb.Message_NNAR_READ:
		// 	nnar.rid++
		// 	nnar.acks = 0
		// 	nnar.readlist = make([]int32, len(nnar.state.Processes))
		// 	nnar.reading = true

		// // When receiving beb broadcast, forward to the beb broadcast perfect link.
		// case pb.Message_BEB_BROADCAST:

		// 	// Trigger a perfect link send to all processes.
		// 	for _, proc := range nnar.state.Processes {
		// 		plsend := pb.Message{
		// 			Type:              pb.Message_PL_SEND,
		// 			FromAbstractionId: msg.GetToAbstractionId(),
		// 			ToAbstractionId:   APP_BEB_PL,
		// 			SystemId:          nnar.state.SystemId,
		// 			MessageUuid:       msg.GetMessageUuid(),
		// 			PlSend: &pb.PlSend{
		// 				Message:     msg.GetBebBroadcast().GetMessage(),
		// 				Destination: proc,
		// 			},
		// 		}
		// 		trigger(nnar.state, nnar.queue, &plsend)
		// 	}

		// // When receiving a pl deliver, generate a beb deliver message and forward
		// //it to app.
		// case pb.Message_PL_DELIVER:
		// 	bebdeliver := pb.Message{
		// 		Type:              pb.Message_BEB_DELIVER,
		// 		FromAbstractionId: msg.GetToAbstractionId(),
		// 		ToAbstractionId:   APP,
		// 		SystemId:          nnar.state.SystemId,
		// 		MessageUuid:       msg.GetMessageUuid(),
		// 		BebDeliver: &pb.BebDeliver{
		// 			Sender:  msg.GetPlDeliver().GetSender(),
		// 			Message: msg.GetPlDeliver().GetMessage(),
		// 		},
		// 	}
		// 	trigger(nnar.state, nnar.queue, &bebdeliver)
		// }
	}
	return nil
}
