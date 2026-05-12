package consumer

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

func NewConsumer(url, queueName string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()

		return nil, err
	}

	return &Consumer{
		conn:    conn,
		channel: ch,
		queue:   queueName,
	}, nil
}

func (c *Consumer) Start(handler func(body []byte) error) error {
	msgs, err := c.channel.Consume(
		c.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			if err := handler(d.Body); err != nil {
				log.Println("Handler error, message will be retried:", err)
				continue
			}
			d.Ack(false)
		}
	}()

	return nil
}

func (c *Consumer) Close() {
	c.channel.Close()
	c.conn.Close()
}
