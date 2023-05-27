package main

import (
	"azdemo/pgdb"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := pgdb.Connect()
	if err != nil {
		return fmt.Errorf("connecting to db: %s", err.Error())
	}

	client, err := azservicebus.NewClientFromConnectionString(os.Getenv("SB_READ_ENDPOINT"), nil)
	if err != nil {
		return fmt.Errorf("connecting to service bus: %s", err.Error())
	}

	receiver, err := client.NewReceiverForQueue("orders", &azservicebus.ReceiverOptions{
		ReceiveMode: azservicebus.ReceiveModePeekLock,
	})
	if err != nil {
		return fmt.Errorf("creating queue receiver: %s", err.Error())
	}
	defer receiver.Close(ctx)

	for {
		messages, err := receiver.ReceiveMessages(ctx, 1, nil)
		if err != nil {
			fmt.Printf("Did not receive message properly: %s\n", err.Error())
		}

		for _, message := range messages {
			body := message.Body
			order := binary.BigEndian.Uint32(body)

			fmt.Printf("processing order %d\n", order)
			time.Sleep(5 * time.Second)
			err := db.Model(&pgdb.Order{}).Where("id = ?", order).Update("status", pgdb.READY).Error
			if err != nil {
				fmt.Printf("Could not update order %d: %s", order, err.Error())
				receiver.AbandonMessage(ctx, message, nil)
			}
			fmt.Printf("order %d processed\n", order)

			err = receiver.CompleteMessage(ctx, message, nil)
			if err != nil {
				fmt.Printf("Could not complete message: %s", err.Error())
			}
		}
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error running worker: %s", err.Error())
		os.Exit(1)
	}
}
