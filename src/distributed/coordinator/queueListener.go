package coordinator

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/streadway/amqp"

	"distributed-example/src/distributed/dataTransferObject"
	"distributed-example/src/distributed/utils"
)

const URL = "amqp://guest:guest@localhost:5672"

type QueueListener struct {
	conn            *amqp.Connection
	ch              *amqp.Channel
	sources         map[string]<-chan amqp.Delivery
	eventAggregator *EventAggregator
}

func NewQueueListener() *QueueListener {
	ql := QueueListener{
		sources:         make(map[string]<-chan amqp.Delivery),
		eventAggregator: NewEventAggregator(),
	}
	ql.conn, ql.ch = utils.GetChannel(URL)
	return &ql
}

func (ql *QueueListener) ListenForNewSource() {
	q := utils.GetQueue("", ql.ch)
	ql.ch.QueueBind(
		q.Name,
		"",
		"amq.fanout",
		false,
		nil,
	)

	msgs, _ := ql.ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	ql.DiscoverSensors()

	for msg := range msgs {
		sourceChan, _ := ql.ch.Consume(
			string(msg.Body),
			"",
			true,
			false,
			false,
			false,
			nil,
		)

		if ql.sources[string(msg.Body)] == nil {
			ql.sources[string(msg.Body)] = sourceChan
			go ql.AddListener(sourceChan)
		}
	}
}

func (queueListener *QueueListener) AddListener(msgs <-chan amqp.Delivery) {
	fmt.Println("here")
	for msg := range msgs {
		r := bytes.NewReader(msg.Body)
		d := gob.NewDecoder(r)
		sensorData := new(dataTransferObject.SensorMessage)
		d.Decode(sensorData)

		fmt.Printf("Reading mesage: %v \n", sensorData)

		eventData := EventData{
			Name:      sensorData.Name,
			Timestamp: sensorData.Timestamp,
			Value:     sensorData.Value,
		}
		queueListener.eventAggregator.PublishEvent("MessageReceived"+msg.RoutingKey, eventData, )

	}
}

func (queueListener *QueueListener) DiscoverSensors() {
	_ = queueListener.ch.ExchangeDeclare(
		utils.SensorDiscoveryExchange,
		"fanout",
		false,
		false,
		false,
		false,
		nil,
	)

	queueListener.ch.Publish(
		utils.SensorDiscoveryExchange,
		"",
		false,
		false,
		amqp.Publishing{},
	)
}
