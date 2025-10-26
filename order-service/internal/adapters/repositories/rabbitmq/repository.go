package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
	"wheres-my-pizza/order-service/internal/core/domain"
	"wheres-my-pizza/order-service/internal/infrastructure/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type publisher struct {
	Conn      *amqp.Connection
	Channel   *amqp.Channel
	url       string
	ctx       context.Context
	ctxCansel context.CancelFunc
}

func NewRabbitMQRepository(config *config.RabbitMQ, ctxMain context.Context) *publisher {
	ctx, ctxCansel := context.WithCancel(ctxMain)
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
		Conn:      conn,
		Channel:   ch,
		url:       url,
		ctx:       ctx,
		ctxCansel: ctxCansel,
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
	if err != nil {
		return fmt.Errorf("Failed to declare an exchange")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	message, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

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
	if err != nil {
		return fmt.Errorf("Failed to publish a message")
	}

	return nil
}

func (publisher *publisher) Reconnect() {
	go func() {
		for {
			conn, err := amqp.Dial(publisher.url)
			if err == nil {
				ch, err := conn.Channel()
				if err == nil {
					publisher.Conn = conn
					publisher.Channel = ch
					break
				}
			}

			if err != nil {
				select {
				case <-time.After(5 * time.Second):
				case <-publisher.ctx.Done():
					return
				}
			}
		}
	}()

}

func (publisher *publisher) IsClosed() (bool, bool) {
	return publisher.Channel.IsClosed(), publisher.Conn.IsClosed()
}

func (publisher *publisher) Close() {
	publisher.ctxCansel()

	publisher.Channel.Close()
	publisher.Conn.Close()

	log.Println("Closed rabbitmq channel, connection")
}
