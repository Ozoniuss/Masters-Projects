package main

import (
	"encoding/binary"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func waiter(wid int) {
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

	q, err := ch.QueueDeclare(
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

	var forever chan struct{}

	go func() {
		for d := range msgs {
			oid := binary.BigEndian.Uint32(d.Body)
			fmt.Printf("[waiter %d] Order with id %d ready\n", wid, oid)
			Orders.ChangeOrderStatus(oid, ORDER_READY)
			// Server-sent events notification
			d.Ack(false)
		}
	}()
	<-forever
}
