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
	"github.com/collinewait/pp-monitoring-system/src/distributed/qutils"
	"github.com/streadway/amqp"
)

var url = "amqp://guest:guest@localhost:5672"

var name = flag.String("name", "sensor", "name of the sensor")
var freq = flag.Uint("freq", 5, "update frequency in cycles/sec")
var max = flag.Float64("max", 5., "maximum value for generated readings")
var min = flag.Float64("min", 1., "minimum value for generated readings")
var stepSize = flag.Float64("step", 0.1, "maximum allowed change per measure")

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

var value float64
var nom float64

func main() {
	flag.Parse()

	value = r.Float64()*(*max-*min) + *min // creates a random value between max and min starting somewhere within permissible range
	nom = (*max-*min)/2 + *min             // holds the nominal value of the sensor

	conn, ch := qutils.GetChannel(url)
	defer conn.Close()
	defer ch.Close()

	dataQueue := qutils.GetQueue(*name, ch, false) // we need to do this step even when we are not going to writing to the queue directly. Declaring the queue here we can be sure that rabbit has set it up properly and it will be ready for us to use

	publishQueueName(ch)

	discoveryQueue := qutils.GetQueue("", ch, true)
	ch.QueueBind(
		discoveryQueue.Name,
		"",
		qutils.SensorDiscoveryExchange,
		false,
		nil,
	)

	go listenForDiscoveryRequests(discoveryQueue.Name, ch)

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
		enc = gob.NewEncoder(buf) // the encoder needs to be reused every time it is going to be reused
		enc.Encode(reading)

		msg := amqp.Publishing{
			Body: buf.Bytes(),
		}

		ch.Publish(
			"", // using the default exchange
			dataQueue.Name,
			false,
			false,
			msg,
		)
		log.Printf("Reading sent. Value: %v\n", value)
	}
}

func listenForDiscoveryRequests(name string, ch *amqp.Channel) {
	msgs, _ := ch.Consume(
		name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	for range msgs {
		publishQueueName(ch)
	}
}

func publishQueueName(ch *amqp.Channel) {
	msg := amqp.Publishing{Body: []byte(*name)}
	ch.Publish(
		"amq.fanout",
		"",
		false,
		false,
		msg,
	)
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
