package main

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

const exchange = "go_exchange"

func main() {
	// create connection
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	// create channel
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	// declare exchange
	err = ch.ExchangeDeclare(
		exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}
	// subscribe messages
	go subscribe(conn, exchange)
	go subscribe(conn, exchange)
	// publish messages
	i := 0
	for {
		i++
		err := ch.Publish(
			exchange,
			"",
			false,
			false,
			amqp.Publishing{
				Body: []byte(fmt.Sprintf("message %d", i)),
			},
		)
		if err != nil {
			fmt.Println(err.Error())
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func subscribe(conn *amqp.Connection, ex string) {
	// create channel
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()
	// create queue
	q, err := ch.QueueDeclare(
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}
	defer ch.QueueDelete(
		q.Name,
		false,
		false,
		false,
	)
	// bind queue
	err = ch.QueueBind(
		q.Name,
		"",
		ex,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}
	// consume message
	consume("c", ch, q.Name)
}

func consume(consumer string, ch *amqp.Channel, q string) {
	// consume message
	msgs, err := ch.Consume(
		q,
		consumer,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}
	// print message
	for msg := range msgs {
		fmt.Printf("%s: %s\n", consumer, msg.Body)
	}
}
