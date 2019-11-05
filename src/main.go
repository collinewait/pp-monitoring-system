package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func main() {
	server()
}

func server() {
	conn, ch, q := getQueue()
	defer conn.Close()
	defer ch.Close()

	msg := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("Hello RabbitMQ"),
	}
	for {
		ch.Publish("", q.Name, false, false, msg)
	}
}

func getQueue() (*amqp.Connection, *amqp.Channel, *amqp.Queue) {
	conn, err := amqp.Dial("amqp://guest@localhost:5672")
	failOnError(err, "Failed to connect to RabbitMQ")
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	q, err := ch.QueueDeclare("hello",
		false, // durable bool. determins if messages should be saved on the disk when they're added to the queue
		false, // autoDelete bool. tells rabbitmq what to do with messages on the queue that don't have active consumers
		false, // exclusive bool. allows us to set this queue to be only accessible from the connection that requests it
		false, // noWait bool,
		nil)   // args amqp.Table)
	failOnError(err, "Failed to declare a queue")

	return conn, ch, &q
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
