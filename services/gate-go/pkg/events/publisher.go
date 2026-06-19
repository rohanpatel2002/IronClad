package events

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher sends events to RabbitMQ.
type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
}

// NewPublisher creates a new RabbitMQ publisher.
func NewPublisher() (*Publisher, error) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	q, err := ch.QueueDeclare(
		"scoring_events", // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare a queue: %w", err)
	}

	return &Publisher{
		conn:    conn,
		channel: ch,
		queue:   q.Name,
	}, nil
}

// PublishScoringEvent sends a scoring request to the queue.
func (p *Publisher) PublishScoringEvent(ctx context.Context, event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return p.channel.PublishWithContext(ctx,
		"",      // exchange
		p.queue, // routing key
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

// Close closes the connection and channel.
func (p *Publisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}
