package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elokanugrah/fleet-management-service/internal/domain"
	"github.com/elokanugrah/fleet-management-service/pkg/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

type GeofencePublisher interface {
	Publish(ctx context.Context, event domain.GeofenceEvent) error
}

type geofencePublisher struct {
	conn *rabbitmq.Connection
}

func NewGeofencePublisher(conn *rabbitmq.Connection) GeofencePublisher {
	return &geofencePublisher{conn: conn}
}

func (p *geofencePublisher) Publish(ctx context.Context, event domain.GeofenceEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("publisher.Publish marshal: %w", err)
	}

	err = p.conn.Channel.PublishWithContext(ctx,
		rabbitmq.ExchangeName, // exchange
		rabbitmq.RoutingKey,   // routing key
		false,                 // mandatory
		false,                 // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // survive broker restart
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("publisher.Publish: %w", err)
	}

	log.Printf("[RabbitMQ] Geofence event published for vehicle %s", event.VehicleID)
	return nil
}
