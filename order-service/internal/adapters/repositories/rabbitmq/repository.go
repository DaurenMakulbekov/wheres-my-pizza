package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"wheres-my-pizza/order-service/internal/core/domain"
	"wheres-my-pizza/order-service/internal/infrastructure/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type publisher struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQRepository(config *config.RabbitMQ) *publisher {
	var url = fmt.Sprintf("amqp://%s:%s@%s:%s/", config.User, config.Password, config.Host, config.Port)

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	return &publisher{
		Conn:    conn,
		Channel: ch,
	}
}

func failOnError(err error, message string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v", message, err)
	}
}

func (publisher *publisher) Publish(order domain.Order) error {
	var err = publisher.Channel.ExchangeDeclare(
		"orders_topic", // name
		"topic",        // type
		true,           // durable
		false,          // auto deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message, err := json.Marshal(order)
	failOnError(err, "Failed to encode")

	var priority = strconv.Itoa(order.Priority)

	err = publisher.Channel.PublishWithContext(ctx,
		"orders_topic",                     // exchange
		"kitchen."+order.Type+"."+priority, // routing key
		false,                              // mandatory
		false,                              // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(message),
		},
	)
	failOnError(err, "Failed to publish a message")

	return nil
}

func (publisher *publisher) Close() {
	publisher.Conn.Close()
	publisher.Channel.Close()
}
