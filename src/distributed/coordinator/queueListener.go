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
	conn    *amqp.Connection
	ch      *amqp.Channel
	sources map[string]<-chan amqp.Delivery
}

func NewQueueListener() *QueueListener {
	ql := QueueListener{
		sources: make(map[string]<-chan amqp.Delivery),
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

func (ql *QueueListener) AddListener(msgs <-chan amqp.Delivery) {
	fmt.Println("here")
	for msg := range msgs {
		r := bytes.NewReader(msg.Body)
		d := gob.NewDecoder(r)
		sd := new(dataTransferObject.SensorMessage)
		d.Decode(sd)

		fmt.Printf("Reading mesage: %v \n", sd)
	}
}
