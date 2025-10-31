package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
	"wheres-my-pizza/kitchen-worker/internal/core/domain"
	"wheres-my-pizza/kitchen-worker/internal/infrastructure/config"

	amqp "github.com/rabbitmq/amqp091-go"
)

type consumer struct {
	Conn      *amqp.Connection
	Channel   *amqp.Channel
	url       string
	ctx       context.Context
	ctxCansel context.CancelFunc
}

func NewRabbitMQRepository(config *config.RabbitMQ, ctxMain context.Context) *consumer {
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

	return &consumer{
		Conn:      conn,
		Channel:   ch,
		url:       url,
		ctx:       ctx,
		ctxCansel: ctxCansel,
	}
}

func (consumer *consumer) ReadMessages(orderTypes []string, prefetch int, out chan string, m map[string]chan bool) {
	for {
		var wg sync.WaitGroup

		for i := range orderTypes {
			var done = make(chan bool)
			m[orderTypes[i]] = done

			var err = consumer.Channel.ExchangeDeclare(
				"orders_topic", // name
				"topic",        // type
				true,           // durable
				false,          // auto deleted
				false,          // internal
				false,          // no-wait
				nil,            // arguments
			)
			if err != nil {
				//return fmt.Errorf("Failed to declare an exchange")
			}

			q, err := consumer.Channel.QueueDeclare(
				"kitchen_"+orderTypes[i]+"_queue", // name
				true,                              // durable
				false,                             // delete when unused
				false,                             // exclusive
				false,                             // no-wait
				nil,                               // arguments
			)
			if err != nil {
				//return fmt.Errorf("Failed to declare a queue")
			}

			err = consumer.Channel.QueueBind(
				q.Name,                           // queue name
				"kitchen."+orderTypes[i]+"."+"*", // routing key
				"orders_topic",                   // exchange
				false,
				nil,
			)
			if err != nil {
				//return fmt.Errorf("Failed to bind a queue")
			}

			err = consumer.Channel.Qos(
				prefetch, // prefetch count
				0,        // prefetch size
				false,    // global
			)
			if err != nil {
				//return fmt.Errorf("Failed to set QoS")
			}

			messages, err := consumer.Channel.Consume(
				q.Name, // queue
				"",     // consumer
				false,  // auto-ack
				false,  // exclusive
				false,  // no-local
				false,  // no-wait
				nil,    // args
			)
			if err != nil {
				//return fmt.Errorf("Failed to register a consumer")
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(done)

				for d := range messages {
					out <- string(d.Body)

					select {
					case requeue := <-done:
						if requeue {
							d.Nack(true, true)
						} else {
							d.Ack(false)
						}
					case <-consumer.ctx.Done():
						return
					}
				}
			}()
		}

		wg.Wait()
		consumer.Reconnect()
	}
}

func (consumer *consumer) PublishStatusUpdate(message domain.Message) error {
	var err = consumer.Channel.ExchangeDeclare(
		"notifications_fanout", // name
		"fanout",               // type
		true,                   // durable
		false,                  // auto deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		return fmt.Errorf("Failed to declare an exchange")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	err = consumer.Channel.PublishWithContext(ctx,
		"notifications_fanout", // exchange
		"",                     // routing key
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(msg),
		},
	)
	if err != nil {
		return fmt.Errorf("Failed to publish a message")
	}

	return nil
}

func (consumer *consumer) Reconnect() {
	for {
		conn, err := amqp.Dial(consumer.url)
		if err == nil {
			ch, err := conn.Channel()
			if err == nil {
				consumer.Conn = conn
				consumer.Channel = ch
				break
			}
		}

		if err != nil {
			select {
			case <-time.After(5 * time.Second):
			case <-consumer.ctx.Done():
				return
			}
		}
	}
}

func (consumer *consumer) IsClosed() (bool, bool) {
	return consumer.Channel.IsClosed(), consumer.Conn.IsClosed()
}

func (consumer *consumer) Close() {
	consumer.ctxCansel()

	consumer.Channel.Close()
	consumer.Conn.Close()

	log.Println("Closed rabbitmq channel, connection")
}
