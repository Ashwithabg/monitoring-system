package main

import (
	"flag"
	"log"
	"math/rand"
	"strconv"
	"time"
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
	numberOfmsPerCycle := strconv.Itoa(1000 / int(*freq)) // 5 cycles/sec => 200 ms/cycle
	dur, _ := time.ParseDuration(numberOfmsPerCycle + "ms")
	signal := time.Tick(dur)

	for range signal {
		calcValue()
		log.Printf("Reading sent. value: %v", value)
	}
}

func calcValue() {
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
