package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/collinewait/pp-monitoring-system/src/distributed/dto"
)

var name = flag.String("name", "sensor", "name of the sensor")
var freq = flag.Uint("freq", 5, "update frequency in cycles/sec")
var max = flag.Float64("max", 5., "maximum value for generated readings")
var min = flag.Float64("min", 1., "minimum value for generated readings")
var stepSize = flag.Float64("step", 0.1, "maximum allowed change per measure")

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

var value = r.Float64()*(*max-*min) + *min // creates a random value between max and min starting somewhere within permissible range
var nom = (*max-*min)/2 + *min             // holds the nominal value of the sensor

func main() {
	flag.Parse()

	/*
		duration object(dur) describes the time between each signal
		provide the milliseconds per cycle instead of cycles persecond
		that we started with. i.e 5 cycles / sec will be converted to
		200 milliseconds per cycle
	*/
	dur, _ := time.ParseDuration(strconv.Itoa(1000/int(*freq)) + "ms")

	signal := time.Tick(dur) // creates a channel that gets triggered at regular intervals that equals the duration created

	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)

	for range signal {
		calcValue()
		reading := dto.SensorMessage{
			Name:      *name,
			Value:     value,
			Timestamp: time.Now(),
		}
		buf.Reset()
		enc.Encode(reading)
		log.Printf("Reading sent. Value: %v\n", value)
	}
}

func calcValue() {
	var maxStep, minStep float64

	if value < nom {
		maxStep = *stepSize
		minStep = -1 * *stepSize * (value - *min) / (nom - *min) // scaled down depending on how close we are to the minimum value of the range
	} else {
		maxStep = *stepSize * (*max - value) / (*min - nom)
		minStep = -1 * *stepSize
	}

	// generate the random value with in the range and add that to the current value
	value += r.Float64()*(maxStep-minStep) + minStep
}
