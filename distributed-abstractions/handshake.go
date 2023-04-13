package main

import (
	"fmt"
	"hw/log"
	pb "hw/protobuf"
	"hw/state"
	"net"
	"net/netip"

	"github.com/google/uuid"
)

/*	Step 1 and 2 of sequence diagram*/

func handshake(state *state.ProcState, host string, listeningPort, index int) error {

	d := net.Dialer{
		LocalAddr: net.TCPAddrFromAddrPort(
			netip.AddrPortFrom(netip.AddrFrom4([4]byte{127, 0, 0, 1}), uint16(listeningPort))),
	}

	// Connect to the fucking hub.
	conn, err := d.Dial(TCP, HUB_ADDRESS)
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
			return fmt.Errorf("could not marshal initialization message: %w", err)
		}

		n, err := conn.Write(msg)
		if err != nil {
			return fmt.Errorf("could not send initialization message: %w", err)
		}
		log.Printf("sent process registration: %d bytes\n", n)
	}

	conn.Close()

	// Wait for confirmation message.
	lis, err := net.Listen(TCP, fmt.Sprintf("127.0.0.1:%d", listeningPort))
	if err != nil {
		return fmt.Errorf("listening failed on the local network: %w", err)
	}
	defer lis.Close()

	client, err := lis.Accept()
	if err != nil {
		return fmt.Errorf("could not accept connection: %w", err)
	}
	defer client.Close()

	// Read the incoming message.
	var received = make([]byte, MAX_MSG_SIZE)
	mlen, err := client.Read(received)

	if err != nil {
		return fmt.Errorf("could not read initialization message from hub: %w", err)
	}

	var m pb.Message
	err = pb.UnmarshalMsg(received[:mlen], &m)
	if err != nil {
		return fmt.Errorf("could not unmarshal initialization message from hub: %s", err)
	}

	if m.GetType() != pb.Message_NETWORK_MESSAGE {
		return fmt.Errorf("did not receive network message during initialization, got %v", m.GetType())
	}
	log.Printf("got ProcInitializeSystem message: %+v\n", &m)

	msg := m.GetNetworkMessage().GetMessage()
	if msg.GetType() != pb.Message_PROC_DESTROY_SYSTEM && msg.GetType() != pb.Message_PROC_INITIALIZE_SYSTEM {
		return fmt.Errorf("received invalid message type during initialization: %v", m.GetType())
	}

	state.SystemId = msg.SystemId
	state.Processes = make([]*pb.ProcessId, 0, len(msg.GetProcInitializeSystem().GetProcesses()))

	for _, pid := range msg.GetProcInitializeSystem().GetProcesses() {
		state.Processes = append(state.Processes, pid)
		if pid.Host == host && pid.Port == int32(listeningPort) {
			if state.CurrentProcId != nil {
				panic("current process already identified")
			}
			state.CurrentProcId = pid
		}
	}

	log.Println("handshake complete")

	return nil
}
