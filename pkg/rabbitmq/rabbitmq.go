package rabbitmq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeName = "fleet.events"
	QueueName    = "geofence_alerts"
	RoutingKey   = "geofence_alerts"
)

type Connection struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewConnection(url string) *Connection {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		ExchangeName, // name
		"direct",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	// Declare queue
	_, err = ch.QueueDeclare(
		QueueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Bind queue to exchange
	err = ch.QueueBind(
		QueueName,    // queue name
		RoutingKey,   // routing key
		ExchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind queue: %v", err)
	}

	log.Println("RabbitMQ connected, exchange and queue ready")
	return &Connection{Conn: conn, Channel: ch}
}

func (c *Connection) Close() {
	if c.Channel != nil {
		c.Channel.Close()
	}
	if c.Conn != nil {
		c.Conn.Close()
	}
}
