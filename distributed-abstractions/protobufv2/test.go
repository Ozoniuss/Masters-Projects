package protobufv2

// Sample broadcast flow

func flow() {

	// Hub delivers broadcast

	// Essentially this says: process that receives this wants to broadcast
	// this message
	appbc := NetworkMessage{
		SenderHost:          "random",
		SenderListeningPort: 5000,
		ToLink:              "app.pl",
		Message: &Message{
			MessageUuid: "irrelevant",
			// Ignored in a network message, next layer determined based on
			// msg type.
			// FromAbstractionId: "app",
			// ToAbstractionId:   "app.beb",
			SystemId: "info",
			Content: &Message_AppBroadcast{
				AppBroadcast: &AppBroadcast{
					// Note that this is not necessary, check only if value is
					// defined or not, or use google's wrappers.
					Value: &Value{
						Defined: true,
						V:       100,
					},
				},
			},
		},
	}

	// Which generates the pl deliver with the broadcast content sent by the app
	apppldeliver := Message{
		MessageUuid:       "random",
		SystemId:          "info",
		FromAbstractionId: "app.pl",
		ToAbstractionId:   "app",
		//Content:           appbc.Message.Content,
		Content: &Message_PlDeliver{
			PlDeliver: &PlDeliver{
				Sender:  nil, // whomever sent the message
				Content: appbc.GetMessage().GetPlDeliver().Content,
			},
		},
	}

	// It is possible to check the content type in the handler
	// and based on the content type decide on what to do
	switch apppldeliver.GetContent().(type) {
	case *Message_AppBroadcast:
		ok := true
	}

	// This receives an appbroadcast content,
	// Which in turn generates the beb broadcast event
	beb := Message{
		MessageUuid:       "random2",
		SystemId:          "info",
		FromAbstractionId: apppldeliver.ToAbstractionId,
		ToAbstractionId:   "app.beb",
		// Message type is changed to beb broadcast
		Content: &Message_BebBroadcast{
			BebBroadcast: &BebBroadcast{
				Content: &BebBroadcast_Value{
					Value: apppldeliver.GetAppBroadcast().Value,
				},
			},
		},
	}

	// Which then goes to the perfect link
	bebplsend := Message{
		MessageUuid:       "random3", // this has to be the same for all messages
		SystemId:          beb.SystemId,
		FromAbstractionId: beb.ToAbstractionId,
		ToAbstractionId:   "app.beb.pl",
		Content: &Message_PlSend{
			PlSend: &PlSend{
				Destination: nil,                          // each process is a destination
				Content:     beb.GetPlSend().GetContent(), // the beb content is sent over the wire
			},
		},
	}

	// In response to PL send every process then creates

	sendovernetwork := NetworkMessage{
		SenderHost:          "own_host",
		SenderListeningPort: 6969,
		ToLink:              bebplsend.ToAbstractionId,

		// From and To abstraction ID are no longer relevant here
		Message: &bebplsend,
	}

	switch sendovernetwork.GetMessage().GetContent().(type) {
	case *Message_BebBroadcast:
		ok := true
	}

	recvbebbroadcast := Message{
		MessageUuid:       sendovernetwork.Message.MessageUuid, // matters because its broadcast
		SystemId:          sendovernetwork.Message.SystemId,
		FromAbstractionId: sendovernetwork.ToLink,
		// Based on the link you will decide
		ToAbstractionId: "app.beb",
		Content: &Message_PlDeliver{
			PlDeliver: &PlDeliver{
				Sender:  nil, // whomever sent it
				Content: sendovernetwork.GetMessage().GetPlDeliver().GetContent(),
			},
		},
	}

}
