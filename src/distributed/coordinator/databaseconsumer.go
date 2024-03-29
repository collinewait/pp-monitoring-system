package coordinator

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/collinewait/pp-monitoring-system/src/distributed/dto"
	"github.com/collinewait/pp-monitoring-system/src/distributed/qutils"
	"github.com/streadway/amqp"
)

const maxRate = 5 * time.Second

type DatabaseConsumer struct {
	er      EventRaiser
	conn    *amqp.Connection
	ch      *amqp.Channel
	queue   *amqp.Queue
	sources []string
}

func NewDatabaseConsumer(er EventRaiser) *DatabaseConsumer {
	dc := DatabaseConsumer{
		er: er,
	}
	/*
		for a production solution, refactor to allow the connection
		to be reused between the QueueListener and the DatabseConsumer
	*/
	dc.conn, dc.ch = qutils.GetChannel(url)
	dc.queue = qutils.GetQueue(
		qutils.PersistReadingsQueue,
		dc.ch,
		false,
	)
	dc.er.AddListener("DataSourceDiscovered", func(eventData interface{}) {
		dc.SubscribeToDataEvent(eventData.(string))
	})

	return &dc
}

func (dc *DatabaseConsumer) SubscribeToDataEvent(eventName string) {
	for _, v := range dc.sources {
		if v == eventName {
			// no need to worry about this event since we are already listening for it
			return
		}
	}
	dc.er.AddListener("MessageReceived_"+eventName, func() func(interface{}) {
		prevTime := time.Unix(0, 0)

		buf := new(bytes.Buffer)

		return func(eventData interface{}) {
			ed := eventData.(EventData)
			if time.Since(prevTime) > maxRate {
				prevTime = time.Now()

				sm := dto.SensorMessage{
					Name:      ed.Name,
					Value:     ed.Value,
					Timestamp: ed.Timestamp,
				}

				buf.Reset()

				enc := gob.NewEncoder(buf)
				enc.Encode(sm)

				msg := amqp.Publishing{
					Body: buf.Bytes(),
				}

				dc.ch.Publish(
					"",                          // exchange string
					qutils.PersistReadingsQueue, // key string
					false,                       // mandatory bool
					false,                       // immediate bool
					msg,                         // amqp.Publishing
				)
			}
		}
	}())
}
