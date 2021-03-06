package utils

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

const SensorDiscoveryExchange = "SensorDiscovery"

func GetChannel(url string) (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(url)
	failOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open channel")

	return conn, ch
}

func GetQueue(name string, ch *amqp.Channel) (*amqp.Queue) {
	q, err := ch.QueueDeclare(name,
		false,
		false,
		false,
		false,
		nil)
	failOnError(err, "Failed to declare queue")

	return &q
}


func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s:%s", msg, err))
	}
}

