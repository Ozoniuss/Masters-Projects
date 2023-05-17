package main

import (
	"context"
	"encoding/binary"
	"flag"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

/*
	Note that worker has no connection to the database in this
	context.
*/

const (
	ORDER_QUEUE = "orders"
	EXCHANGE    = ""
	READY_QUEUE = "ready"
)

func main() {

	var workerno uint

	flag.UintVar(&workerno, "workers", 0, "Specify number of workers")
	flag.Parse()
	fmt.Printf("[worker %d] Initializing worker...\n", workerno)

	conn, err := amqp.Dial("amqp://fast:food@localhost:5672/fastfood")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	// Put the channel in confirm mode.
	ch.Confirm(false)

	// READY_QUEUE to which messages are published when an order is ready.
	_, err = ch.QueueDeclare(
		READY_QUEUE, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		panic(err)
	}

	// ORDERS_QUEUE where messages are consumed from
	q, err := ch.QueueDeclare(
		ORDER_QUEUE, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		panic(err)
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Printf("[worker %d] Worker initialized, listening for orders.\n", workerno)

	var forever chan struct{}

	go func() {
		for d := range msgs {
			oid := binary.BigEndian.Uint32(d.Body)

			fmt.Printf("[worker %d] Started preparing order %d\n", workerno, oid)
			time.Sleep(10 * time.Second)
			fmt.Printf("[worker %d] Order with id %d finished\n", workerno, oid)

			// Publish the finished order to the "ready" queue
			confirmation, err := ch.PublishWithDeferredConfirmWithContext(ctx, EXCHANGE,
				READY_QUEUE,
				false,
				false,
				amqp.Publishing{
					DeliveryMode: amqp.Persistent,
					ContentType:  "text/plain",
					Body:         d.Body,
				})
			if err != nil {
				panic(err)
			}
			ok, err := confirmation.WaitContext(ctx)
			if !ok || err != nil {
				// TODO: retry
				panic("could not send message to ready queue")
			}
			d.Ack(false)
		}
	}()
	<-forever
}
