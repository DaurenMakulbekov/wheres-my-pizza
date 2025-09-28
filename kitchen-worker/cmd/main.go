package main

import (
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, message string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v", message, err)
		os.Exit(1)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"task_exchange", // name
		"fanout",        // type
		true,            // durable
		false,           // auto deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,          // queue name
		"",              // routing key
		"task_exchange", // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	messages, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever = make(chan struct{})

	go func() {
		for d := range messages {

			time.Sleep(time.Second)

			fmt.Printf("Received a message: %s\n", d.Body)
			//d.Ack(false)
		}
	}()

	fmt.Println("Waiting for messages. To exit press a CTRL+C")

	<-forever
}
