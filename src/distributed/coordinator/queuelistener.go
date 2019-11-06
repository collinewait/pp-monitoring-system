package coordinator

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/collinewait/pp-monitoring-system/src/distributed/dto"
	"github.com/collinewait/pp-monitoring-system/src/distributed/qutils"
	"github.com/streadway/amqp"
)

const url = "amqp://guest:guest@localhost:5672"

type QueueListener struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	sources map[string]<-chan amqp.Delivery
	ea      *EventAggregator
}

func NewQueueListener() *QueueListener {
	ql := QueueListener{
		sources: make(map[string]<-chan amqp.Delivery),
		ea:      NewEventAggregator(),
	}

	ql.conn, ql.ch = qutils.GetChannel(url)

	return &ql
}

func (ql *QueueListener) DiscoverSensors() {
	ql.ch.ExchangeDeclare(
		qutils.SensorDiscoveryExchange,
		"fanout",
		false,
		false,
		false,
		false,
		nil,
	)

	ql.ch.Publish(
		qutils.SensorDiscoveryExchange,
		"",
		false,
		false,
		amqp.Publishing{},
	)
}

func (ql *QueueListener) ListenForNewSource() {
	q := qutils.GetQueue("", ql.ch, true)
	ql.ch.QueueBind(
		q.Name,       // name string
		"",           // key string
		"amq.fanout", // exchange string
		false,        // noWait,
		nil,          // arg amqp.Table
	)

	msgs, _ := ql.ch.Consume(
		q.Name, // queue string
		"",     // consumer string
		true,   // autoAck bool,
		false,  // excluive bool,
		false,  // noLocal bool
		false,  // noWait bool,
		nil,    // arg amqp.Table
	)

	ql.DiscoverSensors()

	fmt.Println("listening for new sources")

	for msg := range msgs {
		fmt.Println("new source discovered")
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
	for msg := range msgs {
		reader := bytes.NewReader(msg.Body)
		decoder := gob.NewDecoder(reader)
		sensorMessage := new(dto.SensorMessage)
		decoder.Decode(sensorMessage) // populate the sensor message object

		fmt.Printf("Received message: %v\n", sensorMessage)

		ed := EventData{
			Name:      sensorMessage.Name,
			Timestamp: sensorMessage.Timestamp,
			Value:     sensorMessage.Value,
		}

		ql.ea.PublishEvent("MessageReceived_"+msg.RoutingKey, ed)
	}
}
