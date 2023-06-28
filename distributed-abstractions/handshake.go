package main

import (
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/state"
	"net"

	"github.com/google/uuid"
)

/*	Step 1 and 2 of sequence diagram*/

func handshake(state *state.ProcState, host string, listeningPort, index int) error {

	conn, err := net.Dial(TCP, HUB_ADDRESS)
	if err != nil {
		return fmt.Errorf("could not connect to the hub: %w", err)
	}
	defer conn.Close()

	// Create the process registration message.
	hello := pb.Message{
		Type:        pb.Message_NETWORK_MESSAGE,
		MessageUuid: uuid.NewString(),
		NetworkMessage: &pb.NetworkMessage{
			SenderHost:          host,
			SenderListeningPort: int32(listeningPort),
			Message: &pb.Message{
				Type: pb.Message_PROC_REGISTRATION,
				ProcRegistration: &pb.ProcRegistration{
					Owner: OWNER,
					Index: int32(index),
				},
			},
		},
	}

	// Send process registration message to the hub.
	{
		msg, err := pb.MarshalMsg(&hello)
		if err != nil {
			return fmt.Errorf("sending initialization message: %s", err.Error())
		}

		n, err := conn.Write(msg)
		if err != nil {
			return fmt.Errorf("sending initialization message: %s", err.Error())
		}
		log.Printf("[handshake] sent process registration: %d bytes\n\n", n)
	}

	conn.Close()

	// // Wait for confirmation message.
	// lis, err := net.Listen(TCP, fmt.Sprintf("127.0.0.1:%d", listeningPort))
	// if err != nil {
	// 	return fmt.Errorf("listening during handshake: %s", err.Error())
	// }
	// defer lis.Close()

	// client, err := lis.Accept()
	// if err != nil {
	// 	return fmt.Errorf("accepting connections during handshake: %s", err.Error())
	// }
	// defer client.Close()

	// // Read message header
	// var header = make([]byte, 4)
	// _, err = client.Read(header)
	// if err != nil {
	// 	return fmt.Errorf("reading header during handshake: %s", err.Error())
	// }
	// mlen := binary.BigEndian.Uint32(header)

	// // Read the incoming message.
	// var received = make([]byte, mlen)
	// _, err = client.Read(received)
	// if err != nil {
	// 	return fmt.Errorf("reading system init during handshake: %s", err.Error())
	// }

	// var m pb.Message
	// err = pb.UnmarshalMsg(received[:mlen], &m)
	// if err != nil {
	// 	return fmt.Errorf("unmarshaling system init during handshake: %s", err.Error())
	// }

	// log.Printf("[handshake] got {%+v}\n\n", &m)

	// msg := m.GetNetworkMessage().GetMessage()
	// if msg.GetType() != pb.Message_PROC_INITIALIZE_SYSTEM {
	// 	return fmt.Errorf("received invalid message type during initialization: %s", m.GetType())
	// }

	// state.SystemId = msg.SystemId
	// state.Processes = make([]*pb.ProcessId, 0, len(msg.GetProcInitializeSystem().GetProcesses()))

	// for _, pid := range msg.GetProcInitializeSystem().GetProcesses() {
	// 	// Do not register the hub to the process list.
	// 	if pid.Owner == "hub" {
	// 		continue
	// 	}
	// 	state.Processes = append(state.Processes, pid)
	// 	if pid.Host == host && pid.Port == int32(listeningPort) {
	// 		if state.CurrentProcId != nil {
	// 			panic("current process already identified")
	// 		}
	// 		state.CurrentProcId = pid
	// 	}
	// }
	// log.Printf("[handshake] Completed. \n\n")
	return nil
}
