package producer

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

func NewProducer(url, queueName string) (*Producer, error) {
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

	return &Producer{
		conn:    conn,
		channel: ch,
		queue:   queueName,
	}, nil
}

func (p *Producer) Publish(body []byte) error {
	return p.channel.Publish(
		"",
		p.queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *Producer) Close() {
	p.channel.Close()
	p.conn.Close()
}
