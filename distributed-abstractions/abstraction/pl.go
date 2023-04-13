package abstraction

import (
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
	"net"

	"github.com/google/uuid"
)

const MAX_MSG_SIZE = 256 * 256 * 256 * 256

// Pl acts both as a perfect link for receiving messages from the hub, as well
// as the other processes.
type Pl struct {
	state *procstate.ProcState
	queue *queue.Queue
}

func NewPl(state *procstate.ProcState, queue *queue.Queue) *Pl {
	return &Pl{
		state: state,
		queue: queue,
	}
}

func (pl *Pl) Handle(msg *pb.Message) error {

	switch msg.GetType() {

	// App.pl is what actually receives network messages.
	case pb.Message_PL_SEND:

		var PL string
		if msg.ToAbstractionId == APP_BEB_PL {
			PL = APP_BEB_PL
		} else {
			PL = APP_PL
		}

		destName := fmt.Sprintf("%s-%d", msg.GetPlSend().GetDestination().GetOwner(), msg.GetPlSend().GetDestination().GetIndex())

		// Do not use the custom dialer because the perfect link is already
		// listening on its IP address. This means another IP address must
		// be specified since the socket cannot be re-used. The sender
		// information is included in the network message anyway.
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", msg.GetPlSend().GetDestination().GetHost(), msg.GetPlSend().GetDestination().GetPort()))
		if err != nil {
			return fmt.Errorf("%s could not connect to the hub: %w", PL, err)
		}
		defer conn.Close()

		// Create the network message to be delivered to the other processes.
		sendmsg := pb.Message{
			Type:              pb.Message_NETWORK_MESSAGE,
			MessageUuid:       uuid.NewString(),
			FromAbstractionId: msg.ToAbstractionId,
			ToAbstractionId:   APP_BEB_PL,
			SystemId:          pl.state.SystemId,
			NetworkMessage: &pb.NetworkMessage{
				// SenderHost:          pl.state.CurrentProcId.Host, (irrelevant)
				SenderListeningPort: pl.state.CurrentProcId.Port,
				Message:             msg.GetPlSend().GetMessage(),
			},
		}

		log.Printf("[SENDING OVER NETWORK]: {%+v}\n\n", &sendmsg)

		// Marshal the network message and send it to the other processes.
		msgbyte, err := pb.MarshalMsg(&sendmsg)
		if err != nil {
			return fmt.Errorf("%s could not marshal message %+v: %w", PL, &sendmsg, err)
		}

		_, err = conn.Write(msgbyte)
		if err != nil {
			return fmt.Errorf("%s could not send message %+v: %w", PL, &sendmsg, err)
		}

		log.Printf("%s/%s -> %s : {%+v} \n\n", pl.state.SystemId, pl.state.Name(), destName, sendmsg.GetNetworkMessage().GetMessage())

		conn.Close()
	}
	return nil
}

func (pl *Pl) ReadSocket(lis net.Listener) error {

	for {

		var client net.Conn
		// var err error
		// var gotClient = make(chan struct{}, 1)

		// go func() {
		// 	// Accept incoming messages.
		// 	client, err = lis.Accept()
		// 	if err != nil {
		// 		app_pl.logg.Errorf("could not accept incoming connection: %s", err.Error())
		// 		app_pl.state.Quit <- struct{}{}
		// 		//panic(err)
		// 	} else {
		// 		gotClient <- struct{}{}
		// 	}
		// }()

		// select {
		// case <-app_pl.state.Quit:
		// 	break
		// case <-gotClient:
		// }

		client, err := lis.Accept()
		if err != nil {
			panic(err)
		}

		defer client.Close()
		var received = make([]byte, MAX_MSG_SIZE)
		mlen, err := client.Read(received)

		if err != nil {
			return fmt.Errorf("could not read incoming message: %w", err)
		}

		// log.Printf("%s-%d : received bytes %v\n", pl.state.CurrentProcId.Owner, pl.state.CurrentProcId.Index, received[:mlen])

		var msg pb.Message
		err = pb.UnmarshalMsg(received[:mlen], &msg)
		if err != nil {
			return fmt.Errorf("could not unmarshal incoming message: %w", err)
		}

		if msg.GetType() != pb.Message_NETWORK_MESSAGE {
			return fmt.Errorf("did not receive network message, got {%+v}", &msg)
		}

		// Upon receiving a network message, we can trigger a private link
		// deliver. Note that triggering the private link deliver acts similar
		// to an internal event -- it's not the response to some other event,
		// but internally the perfect link detected the reception of the
		// message.
		//
		// Note, as per page 34 of the book, "the deliver indication is
		// triggered by the algorithm implementing the abstraction on a
		// destination process. When this event occurs on a process p for
		// a message m, we say that p delivers m". This transcribes to my
		// program as in "once the message is received and decoded, it is
		// delivered by placing the inner of the network message in a
		// processing queue."
		//
		// For this reason, network messages are processed separately and not
		// via the handle method.

		// Hub doesn't send via perfect link sometimes.
		// sender := "hub"
		// if msg.GetNetworkMessage().GetMessage().GetPlSend() != nil {
		// 	sender = fmt.Sprintf("%s/%d", msg.GetNetworkMessage().GetMessage().GetPlSend().GetDestination().GetOwner(), msg.GetNetworkMessage().GetMessage().GetPlSend().GetDestination().GetIndex())
		// }

		log.Printf("[GOT NETWORK MESSAGE] %s/%s <- : {%v}\n\n", pl.state.SystemId, pl.state.Name(), &msg)

		// Got a command from the hub.
		//
		// Can be either app.pl or app.beb.pl, so take it from the
		// message instead of hardcoding. The network message indicates
		// the actual receiver of this message over the network. The
		// "toAbsractionId" inside the network message indicates the
		// layer where the message was actually destined to go.
		if msg.GetToAbstractionId() == APP_PL {
			deliver := pb.Message{
				Type:              pb.Message_PL_DELIVER,
				FromAbstractionId: APP_PL,
				ToAbstractionId:   msg.GetNetworkMessage().GetMessage().GetToAbstractionId(),
				SystemId:          msg.GetSystemId(),
				PlDeliver: &pb.PlDeliver{
					Message: msg.GetNetworkMessage().GetMessage(),
				},
			}
			// Trigger a perfect link deliver.
			trigger(pl.state, pl.queue, &deliver)
		} else if msg.GetToAbstractionId() == APP_BEB_PL {
			deliver := pb.Message{
				Type:              pb.Message_PL_DELIVER,
				FromAbstractionId: APP_BEB_PL,
				ToAbstractionId:   APP_BEB,
				PlDeliver: &pb.PlDeliver{
					Message: msg.NetworkMessage.Message,
				},
			}
			// Trigger a perfect link deliver.
			trigger(pl.state, pl.queue, &deliver)
		} else {
			log.Printf("received network message that doesn't go to perfect link, but to %v\n\n", msg.GetToAbstractionId())
			pl.state.Quit <- struct{}{}
		}

		client.Close()
	}
}
