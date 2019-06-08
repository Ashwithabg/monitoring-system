package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/streadway/amqp"

	"distributed-example/src/distributed/dataTransferObject"
	"distributed-example/src/distributed/utils"
)

var BrokerInputListenerURL = "amqp://guest:guest@localhost:5672"

//Read parameters from command line
var name = flag.String("name", "sensor", "name of the sensor.(Making it as the routing key. So it should be unique)")
var frequency = flag.Uint("frequency", 5, "update frequency in cycles/sec. (How many datapoints to be generated per second)")

//max and min limit for measurement change
var maximumValue = flag.Float64("maximum", 5., "maximum value for generated readings")
var minimumValue = flag.Float64("minimum", 1., "minimum value for generated readings")
var stepSize = flag.Float64("step", 0.1, "maximum allowable change per measurement")

//Generate data points
var randomNumber = rand.New(rand.NewSource(time.Now().UnixNano()))
var value = randomNumber.Float64()*(*maximumValue-*minimumValue) + *minimumValue
var nominalValueOfSensor = (*maximumValue-*minimumValue)/2 + *minimumValue

func main() {
	flag.Parse()

	connection, channel := utils.GetChannel(BrokerInputListenerURL)
	defer connection.Close()
	defer channel.Close()

	//Uses direct route so only one consumer will receive the data.
	dataQueue := utils.GetQueue(*name, channel)

	msg := amqp.Publishing{Body: []byte(*name)}
	channel.Publish(
		"amq.fanout",
		"",
		false,
		false,
		msg)

	signals := getSignals()
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)

	for range signals {
		buffer.Reset()
		publishMessage(encoder, buffer, channel, dataQueue)
	}
}

func publishMessage(encoder *gob.Encoder, buffer *bytes.Buffer, channel *amqp.Channel, dataQueue *amqp.Queue) {
	calculateValues()

	reading := dataTransferObject.SensorMessage{Name: *name, Value: value, Timestamp: time.Now(),}
	encoder.Encode(reading)
	msg := amqp.Publishing{Body: buffer.Bytes()}
	channel.Publish("", dataQueue.Name, false, false, msg)

	log.Printf("Reading sent. value: %v", value)
}

//Get signal in given frequency.
func getSignals() <-chan time.Time {
	//duration describes the time time between each signal
	numberOfmsPerCycle := strconv.Itoa(1000 / int(*frequency)) // 5 cycles/sec => 200 ms/cycle
	duration, _ := time.ParseDuration(numberOfmsPerCycle + "ms")

	//Create channel that triggers at regular intervals.
	signal := time.Tick(duration)
	return signal
}

//Some random logic to generate the values ti vary in time.
func calculateValues() {
	var maxStep, minStep float64

	if value < nominalValueOfSensor {
		maxStep = *stepSize
		minStep = -1 * *stepSize * (value - *minimumValue) / (nominalValueOfSensor - *minimumValue)
	} else {
		maxStep = *stepSize * (*maximumValue - value) / (*maximumValue - nominalValueOfSensor)
		minStep = -1 * *stepSize
	}

	value += randomNumber.Float64()*(maxStep-minStep) + minStep
}
