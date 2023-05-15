package main

import (
	"context"
	"encoding/binary"
	"fastfood/orders"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func worker(wid int, conn *amqp.Connection) {

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

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

	var forever chan struct{}

	go func() {
		for d := range msgs {
			oid := binary.BigEndian.Uint32(d.Body)

			OrdersDB.ChangeOrderStatus(oid, orders.ORDER_PREPARING)
			fmt.Printf("[worker %d] Started preparing order %d\n", wid, oid)
			time.Sleep(10 * time.Second)
			fmt.Printf("[worker %d] Order with id %d finished\n", wid, oid)

			// Publish the finished order to the "ready" queue
			err = ch.PublishWithContext(ctx, EXCHANGE,
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
			d.Ack(false)
		}
	}()
	<-forever
}
