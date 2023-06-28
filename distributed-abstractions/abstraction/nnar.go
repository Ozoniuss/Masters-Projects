package abstraction

import (
	pb "hw/protobuf"
	"hw/queue"
	procstate "hw/state"
)

type Nnar struct {
	state      *procstate.ProcState
	queue      *queue.Queue
	ts         int       // timestamp
	wr         int       // writerrank
	val        *pb.Value // value
	acks       int       // acknowlledges
	writeval   *pb.Value
	readval    *pb.Value
	rid        int
	readlist   map[int]*pb.NnarInternalValue
	reading    bool
	registerId string
}

func NewNnar(state *procstate.ProcState, queue *queue.Queue, registerId string) *Nnar {
	return &Nnar{
		state:      state,
		queue:      queue,
		ts:         0,
		wr:         0,
		val:        nil,
		writeval:   nil,
		readval:    nil,
		rid:        0,
		readlist:   make(map[int]*pb.NnarInternalValue),
		reading:    false,
		registerId: registerId,
	}
}

func (nnar *Nnar) Handle(msg *pb.Message) error {

	switch msg.GetType() {

	// Handling NNAR_WRITE, received after emitting an APP_WRITE from the hub.
	case pb.Message_NNAR_WRITE:
		nnar.rid++
		nnar.writeval = msg.GetNnarWrite().GetValue()
		// nnar.writeval = &pb.Value{
		// 	Defined: true,
		// 	V: msg.GetNnarWrite().GetValue().GetV(),
		// }
		nnar.acks = 0
		nnar.readlist = make(map[int]*pb.NnarInternalValue)
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

	case pb.Message_NNAR_READ:
		nnar.rid++
		nnar.acks = 0
		nnar.reading = true
		nnar.readlist = make(map[int]*pb.NnarInternalValue)

		bebBroadcast := pb.Message{
			Type:              pb.Message_BEB_BROADCAST,
			FromAbstractionId: msg.GetToAbstractionId(),
			ToAbstractionId:   Next(msg.GetToAbstractionId(), "beb"),
			BebBroadcast: &pb.BebBroadcast{
				Message: &pb.Message{
					Type:              pb.Message_NNAR_INTERNAL_READ,
					FromAbstractionId: msg.GetToAbstractionId(), // app.nnar[x]
					ToAbstractionId:   msg.GetToAbstractionId(), // app.nnar[x]
					NnarInternalRead: &pb.NnarInternalRead{
						ReadId: int32(nnar.rid),
					},
				},
			},
		}
		trigger(nnar.state, nnar.queue, &bebBroadcast)

	// A beb, Deliver event handler is registered for a READ, a WRITE and
	// an ack.
	case pb.Message_BEB_DELIVER:
		switch msg.GetBebDeliver().GetMessage().GetType() {
		case pb.Message_NNAR_INTERNAL_READ:
			plSend := pb.Message{
				Type:              pb.Message_PL_SEND,
				FromAbstractionId: msg.GetToAbstractionId(),
				ToAbstractionId:   Next(msg.GetToAbstractionId(), "pl"),
				PlSend: &pb.PlSend{
					Destination: msg.GetBebDeliver().GetSender(),
					Message: &pb.Message{
						Type:              pb.Message_NNAR_INTERNAL_VALUE,
						FromAbstractionId: msg.GetFromAbstractionId(), // app.nnar[x]
						ToAbstractionId:   msg.GetToAbstractionId(),   // app.nnarp[x]
						NnarInternalValue: &pb.NnarInternalValue{
							ReadId:     int32(nnar.rid),
							Timestamp:  int32(nnar.ts),
							WriterRank: int32(nnar.wr),
							Value:      nnar.val,
						},
					},
				},
			}
			trigger(nnar.state, nnar.queue, &plSend)

		case pb.Message_NNAR_INTERNAL_WRITE:
			nnarwrite := msg.GetBebDeliver().GetMessage().GetNnarInternalWrite()
			readid := nnarwrite.GetReadId()
			newTs := nnarwrite.GetTimestamp()
			newWr := nnarwrite.GetWriterRank()
			newVal := nnarwrite.GetValue()

			if (newTs > int32(nnar.ts)) || (newTs == int32(nnar.ts) && newWr > int32(nnar.wr)) {
				nnar.ts = int(newTs)
				nnar.wr = int(newWr)
				nnar.val = newVal
			}

			plSend := pb.Message{
				Type:              pb.Message_PL_SEND,
				FromAbstractionId: msg.GetToAbstractionId(),
				ToAbstractionId:   Next(msg.GetToAbstractionId(), "pl"),
				PlSend: &pb.PlSend{
					Destination: msg.GetBebDeliver().GetSender(),
					Message: &pb.Message{
						Type:              pb.Message_NNAR_INTERNAL_ACK,
						FromAbstractionId: msg.GetFromAbstractionId(), // app.nnar[x]
						ToAbstractionId:   msg.GetToAbstractionId(),   // app.nnarp[x]
						NnarInternalAck: &pb.NnarInternalAck{
							ReadId: readid,
						},
					},
				},
			}
			trigger(nnar.state, nnar.queue, &plSend)
		}

	// There is a PL deliver for ACK and for VALUE
	case pb.Message_PL_DELIVER:

		switch msg.GetPlDeliver().GetMessage().GetType() {
		case pb.Message_NNAR_INTERNAL_VALUE:
			// For simplification just break here instead of requeing.
			if msg.GetPlDeliver().GetMessage().GetNnarInternalValue().GetReadId() != int32(nnar.rid) {
				break
			}

			nnarInternalValueMsg := msg.GetPlDeliver().GetMessage().GetNnarInternalValue()
			readid := nnarInternalValueMsg.GetReadId()
			newTs := nnarInternalValueMsg.GetTimestamp()
			newWr := nnarInternalValueMsg.GetWriterRank()
			newVal := nnarInternalValueMsg.GetValue()

			senderPort := msg.GetPlDeliver().GetSender().GetPort()

			nnar.readlist[int(senderPort)] = &pb.NnarInternalValue{
				ReadId:     readid,
				Timestamp:  newTs,
				WriterRank: newWr,
				Value:      newVal,
			}

			if len(nnar.readlist) > len(nnar.state.Processes)/2 {
				maxTs := -1
				maxWr := -1

				for _, internalVal := range nnar.readlist {
					if internalVal.GetTimestamp() > int32(maxTs) ||
						(internalVal.GetTimestamp() == int32(maxTs) &&
							internalVal.GetValue().GetV() > nnar.readval.GetV()) {
						nnar.readval = internalVal.GetValue()
						maxTs = int(internalVal.GetTimestamp())
						maxWr = int(internalVal.GetWriterRank())
					}
				}

				nnar.readlist = make(map[int]*pb.NnarInternalValue)
				var nnarInternaWrite *pb.NnarInternalWrite
				if nnar.reading {
					nnarInternaWrite = &pb.NnarInternalWrite{
						ReadId:     nnarInternalValueMsg.GetReadId(),
						Timestamp:  int32(maxTs),
						WriterRank: int32(maxWr),
						Value:      nnar.readval,
					}
				} else {
					nnarInternaWrite = &pb.NnarInternalWrite{
						ReadId:     nnarInternalValueMsg.GetReadId(),
						Timestamp:  int32(maxTs) + 1,
						WriterRank: nnar.state.CurrentProcId.Rank,
						Value:      nnar.writeval,
					}
				}
				bebBroadcast := pb.Message{
					Type:              pb.Message_BEB_BROADCAST,
					FromAbstractionId: msg.GetToAbstractionId(),
					ToAbstractionId:   Next(msg.GetToAbstractionId(), "beb"),
					BebBroadcast: &pb.BebBroadcast{
						Message: &pb.Message{
							Type:              pb.Message_NNAR_INTERNAL_WRITE,
							FromAbstractionId: msg.GetToAbstractionId(),
							ToAbstractionId:   msg.GetToAbstractionId(),
							NnarInternalWrite: nnarInternaWrite,
						},
					},
				}
				trigger(nnar.state, nnar.queue, &bebBroadcast)
			}

		case pb.Message_NNAR_INTERNAL_ACK:
			// For simplification just break here instead of requeing.
			if msg.GetPlDeliver().GetMessage().GetNnarInternalAck().GetReadId() != int32(nnar.rid) {
				break
			}
			nnar.acks++
			if nnar.acks > len(nnar.state.Processes)/2 {
				nnar.acks = 0
				deliver := &pb.Message{}
				if nnar.reading {
					nnar.reading = false
					deliver = &pb.Message{
						Type:              pb.Message_NNAR_READ_RETURN,
						FromAbstractionId: msg.GetToAbstractionId(),
						ToAbstractionId:   Previous(msg.GetToAbstractionId()),
						NnarReadReturn: &pb.NnarReadReturn{
							Value: nnar.readval,
						},
					}
				} else {
					deliver = &pb.Message{
						Type:              pb.Message_NNAR_WRITE_RETURN,
						FromAbstractionId: msg.GetToAbstractionId(),
						ToAbstractionId:   Previous(msg.GetToAbstractionId()),
						NnarWriteReturn:   &pb.NnarWriteReturn{},
					}
				}
				trigger(nnar.state, nnar.queue, deliver)
			}
		}
	}
	return nil
}
