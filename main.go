package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func main() {
	server()
	client()
}

func client() {
	conn, ch, q := getQueue()
	defer conn.Close()
	defer ch.Close()

	msgs, err := ch.Consume(q.Name, "", true,
		false, false, false, nil)

	failOnError(err, "failed to register consumer")

	for msg := range msgs {
		log.Printf("Received message: %s", msg.Body)
	}

}

func server() {
	conn, ch, q := getQueue()
	defer conn.Close()
	defer ch.Close()

	msg := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("Hello rabbit"),
	}

	ch.Publish("", q.Name, false, false, msg)

}

func getQueue() (*amqp.Connection, *amqp.Channel, *amqp.Queue) {
	conn, err := amqp.Dial("amqp://guest@localhost:5672")
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open channel")

	q, err := ch.QueueDeclare("hello",
		false,
		false,
		false,
		false,
		nil)
	failOnError(err, "Failed to declare queue")

	return conn, ch, &q
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s:%s", msg, err))
	}
}
