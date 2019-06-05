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

var URL = "amqp://guest:guest@localhost:5672"
var name = flag.String("name", "sensor", "name of the sensor")
var freq = flag.Uint("freq", 5, "update frequency in cycles/sec")
var maximumValue = flag.Float64("max", 5., "maximum value for generated readings")
var minimumValue = flag.Float64("min", 1., "minimum value for generated readings")
var stepSize = flag.Float64("step", 0.1, "maximum allowable change per measurement")
var randomNumber = rand.New(rand.NewSource(time.Now().UnixNano()))
var value = randomNumber.Float64()*(*maximumValue-*minimumValue) + *minimumValue
var nominalValueOfSensor = (*maximumValue-*minimumValue)/2 + *minimumValue

func main() {
	flag.Parse() //parse command line flags

	connection, channel := utils.GetChannel(URL)
	defer connection.Close()
	defer channel.Close()

	dataQueue := utils.GetQueue(*name, channel)
	signals := getSignals()
	buffer := new(bytes.Buffer)
	encoder := gob.NewEncoder(buffer)

	for range signals {
		calculateValues()
		reading := dataTransferObject.SensorMessage{
			Name:      *name,
			Value:     value,
			Timestamp: time.Now(),
		}

		buffer.Reset()
		encoder.Encode(reading)

		msg := amqp.Publishing{
			Body: buffer.Bytes(),
		}

		channel.Publish("", dataQueue.Name, false, false, msg)
		log.Printf("Reading sent. value: %v", value)
	}
}

func getSignals() <-chan time.Time {
	numberOfmsPerCycle := strconv.Itoa(1000 / int(*freq)) // 5 cycles/sec => 200 ms/cycle
	duration, _ := time.ParseDuration(numberOfmsPerCycle + "ms")
	signal := time.Tick(duration)
	return signal
}

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
