package abstraction

import (
	"encoding/binary"
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
	"net"
	"strings"

	"github.com/google/uuid"
)

const MAX_MSG_SIZE = 256 * 256 * 256 * 256

// Pl acts both as a perfect link for receiving messages from the hub, as well
// as the other processes.
type Pl struct {
	state         *procstate.ProcState
	queue         *queue.Queue
	abstractionId string
	abstractions  *map[string]Abstraction
}

func NewPl(state *procstate.ProcState, queue *queue.Queue, abstractionId string, abstractions *map[string]Abstraction) *Pl {
	return &Pl{
		state:         state,
		queue:         queue,
		abstractions:  abstractions,
		abstractionId: abstractionId,
	}
}

func (pl *Pl) Handle(msg *pb.Message) error {

	switch msg.GetType() {

	// App.pl is what actually receives network messages.
	case pb.Message_PL_SEND:

		destName := fmt.Sprintf("%s-%d", msg.GetPlSend().GetDestination().GetOwner(), msg.GetPlSend().GetDestination().GetIndex())

		// Hack to send things to the hub.
		if msg.GetPlSend().GetDestination().GetOwner() == "hub" {
			destName = "hub"
		}

		// Do not use the custom dialer because the perfect link is already
		// listening on its IP address. This means another IP address must
		// be specified since the socket cannot be re-used. The sender
		// information is included in the network message anyway.
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", msg.GetPlSend().GetDestination().GetHost(), msg.GetPlSend().GetDestination().GetPort()))
		if err != nil {
			return fmt.Errorf("%s could not connect to external process %s: %w", msg.GetPlSend().GetDestination(), pl.abstractionId, err)
		}
		defer conn.Close()

		// Create the network message to be delivered to the other processes.
		sendmsg := pb.Message{
			Type:              pb.Message_NETWORK_MESSAGE,
			MessageUuid:       uuid.NewString(),
			FromAbstractionId: pl.abstractionId,
			ToAbstractionId:   msg.GetToAbstractionId(), // each pl sends to its twin pl
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
			return fmt.Errorf("%s could not marshal message %+v: %w", pl.abstractionId, &sendmsg, err)
		}

		_, err = conn.Write(msgbyte)
		if err != nil {
			return fmt.Errorf("%s could not send message %+v: %w", pl.abstractionId, &sendmsg, err)
		}

		log.Printf("%s/%s -> %s : {%+v} \n\n", pl.state.SystemId, pl.state.Name(), destName, sendmsg.GetNetworkMessage().GetMessage())

		conn.Close()
	}
	return nil
}

func (pl *Pl) ReadSocket(lis net.Listener) error {
	for {
		var client net.Conn
		client, err := lis.Accept()
		if err != nil {
			panic(err)
		}

		defer client.Close()

		var header = make([]byte, 4)
		_, err = client.Read(header)
		if err != nil {
			panic("COULD NOT READ MESSAGE HEADER\n\n")
		}

		// We know this is going to be the size of the message.
		mlen := binary.BigEndian.Uint32(header)

		// Read that many bytes.
		var received = make([]byte, mlen)
		_, err = client.Read(received)

		if err != nil {
			return fmt.Errorf("could not read incoming message: %w", err)
		}

		// log.Printf("%s-%d : received bytes %v\n", pl.state.CurrentProcId.Owner, pl.state.CurrentProcId.Index, received[:mlen])

		var msg pb.Message
		err = pb.UnmarshalMsg(received, &msg)
		if err != nil {
			return fmt.Errorf("could not unmarshal incoming message: %w", err)
		}

		if msg.GetType() != pb.Message_NETWORK_MESSAGE {
			return fmt.Errorf("did not receive network message, got {%+v}", &msg)
		}
		log.Printf("[GOT NETWORK MESSAGE] %s/%s <- : {%v}\n\n", pl.state.SystemId, pl.state.Name(), &msg)

		// Some abstractions have to be created, such as nnar.
		to := msg.GetToAbstractionId()
		parts := strings.Split(to, ".")
		// In some cases a messages first gets to a nnar's perfect link.
		if len(parts) > 1 && parts[0] == "app" && strings.Contains(parts[1], "nnar") {
			registerId := parts[0] + "." + parts[1]
			if _, ok := (*pl.abstractions)[registerId]; !ok {
				nnar := NewNnar(pl.state, pl.queue, registerId)
				RegisterNnar(pl.abstractions, registerId, nnar, pl.state, pl.queue)
			}
		}

		deliver := pb.Message{
			Type: pb.Message_PL_DELIVER,
			// ToAbstractionId of the network message tells which pl gets the
			// message
			FromAbstractionId: msg.GetToAbstractionId(),
			ToAbstractionId:   Previous(msg.GetToAbstractionId()),
			SystemId:          msg.GetSystemId(),
			PlDeliver: &pb.PlDeliver{
				Message: msg.GetNetworkMessage().GetMessage(),
				Sender: &pb.ProcessId{
					Host: msg.GetNetworkMessage().GetSenderHost(),
					Port: msg.GetNetworkMessage().GetSenderListeningPort(),
				},
			},
		}
		trigger(pl.state, pl.queue, &deliver)
		client.Close()
	}
}
