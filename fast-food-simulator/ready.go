package main

import (
	"encoding/binary"
	"fastfood/orders"
	"fmt"
	"time"

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
			oid := binary.BigEndian.Uint32(d.Body)
			fmt.Printf("[waiter] Pickued up order %d ready\n", oid)
			time.Sleep(2 * time.Second)
			fmt.Printf("[waiter] Order with id %d ready\n", oid)
			OrdersDB.ChangeOrderStatus(oid, orders.ORDER_READY)
			// Server-sent events notification
			d.Ack(false)
		}
	}()
	<-forever
}
