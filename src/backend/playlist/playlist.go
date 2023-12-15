package playlist

import (
	"bytes"
	"log"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var conn, _ = amqp.Dial("amqps://singsphere:singsphere123@b-dd7ec1e7-2096-4e6a-9dec-f8a6dc939959.mq.us-east-1.amazonaws.com:5671")

type Playlist struct {
	queue   amqp.Queue
	channel *amqp.Channel
	Songs   chan string
}

func New() *Playlist {
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	// defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		panic(err)
	}

	return &Playlist{
		queue:   q,
		channel: ch,
	}
}

func (playlist *Playlist) Subscribe(topic string, requests chan<- string) {
	log.Printf("subscribe to key: %s\n", topic)
	playlist.channel.QueueBind(
		playlist.queue.Name,
		topic,
		"songs_exchange",
		false,
		nil,
	)
	msgs, err := playlist.channel.Consume(
		playlist.queue.Name, // queue
		"",                  // consumer
		true,                // auto ack
		false,               // exclusive
		false,               // no local
		false,               // no wait
		nil,                 // args
	)
	if err != nil {
		panic(err)
	}
	go func() {
		for d := range msgs {
			log.Printf("[Playlist] enqueue song '%s' to playlist\n", d.Body)
			requests <- strings.ReplaceAll(string(d.Body[:]), "\"", "")
		}
	}()
}

func (playlist *Playlist) Test(topic string) {
	// conn, err := amqp.Dial("amqps://singsphere:singsphere123@b-dd7ec1e7-2096-4e6a-9dec-f8a6dc939959.mq.us-east-1.amazonaws.com:5671")
	// if err != nil {
	// 	panic(err)
	// }
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		topic, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		panic(err)
	}

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		panic("Failed to register a consumer")
	}

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			dotCount := bytes.Count(d.Body, []byte("."))
			t := time.Duration(dotCount)
			time.Sleep(t * time.Second)
			log.Printf("Done")
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
