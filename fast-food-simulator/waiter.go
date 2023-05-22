package main

import (
	"encoding/binary"
	"fastfood/orders"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func waiter(conn *amqp.Connection) {

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
			oid := binary.BigEndian.Uint32(d.Body[:4])

			status := d.Body[4]
			if status == 0 {
				fmt.Printf("[waiter] Order with id %d is in preparation.\n", oid)
				// Notifies the Update to the registered channel, if any.
				OrdersDB.ChangeOrderStatus(oid, orders.ORDER_PREPARING)
			} else {
				fmt.Printf("[waiter] Order with id %d is ready.\n", oid)
				// Notifies the Update to the registered channel, if any.
				OrdersDB.ChangeOrderStatus(oid, orders.ORDER_READY)
			}

			d.Ack(false)
		}
	}()
	<-forever
}
