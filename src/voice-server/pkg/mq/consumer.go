package mq

import (
	"log"
	"os"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	queue     amqp.Queue
	channel   *amqp.Channel
	requestCh chan string
}

// New creates a new consumer in the RabbitMQ instance
func New(topic string, conn *amqp.Connection) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		"",    // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return nil, err
	}

	consumer := &Consumer{
		queue:     q,
		channel:   ch,
		requestCh: make(chan string),
	}
	consumer.subscribe(topic)
	return consumer, nil
}

// subscribe makes this consumer subscribes to the given topic
func (c *Consumer) subscribe(topic string) {
	log.Printf("subscribe to key: %s\n", topic)
	c.channel.QueueBind(
		c.queue.Name,
		topic,
		os.Getenv("MQ_EXCHANGES_NAME"),
		false,
		nil,
	)
	msgs, err := c.channel.Consume(
		c.queue.Name, // queue
		"",           // consumer
		true,         // auto ack
		false,        // exclusive
		false,        // no local
		false,        // no wait
		nil,          // args
	)
	if err != nil {
		panic(err)
	}
	go func() {
		for d := range msgs {
			log.Printf("[Playlist] enqueue song '%s' to playlist\n", d.Body)
			c.requestCh <- strings.ReplaceAll(string(d.Body[:]), "\"", "")
		}
	}()
}
