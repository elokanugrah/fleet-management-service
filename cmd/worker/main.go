package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/elokanugrah/fleet-management-service/config"
	"github.com/elokanugrah/fleet-management-service/internal/domain"
	"github.com/elokanugrah/fleet-management-service/pkg/rabbitmq"
)

func main() {
	cfg := config.Load()

	rmqConn := rabbitmq.NewConnection(cfg.RabbitMQURL)
	defer rmqConn.Close()

	msgs, err := rmqConn.Channel.Consume(
		rabbitmq.QueueName, // queue
		"fleet-worker",     // consumer tag
		false,              // auto-ack (false = manual ack)
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	log.Printf("Worker started, consuming from queue: %s", rabbitmq.QueueName)

	go func() {
		for msg := range msgs {
			processGeofenceEvent(msg.Body)
			// Manual acknowledge
			msg.Ack(false)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Worker shutting down...")
}

func processGeofenceEvent(body []byte) {
	var event domain.GeofenceEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("[Worker] Failed to parse event: %v | raw: %s", err, string(body))
		return
	}

	log.Printf("[Worker] Geofence event received: vehicle=%s event=%s lat=%.6f lng=%.6f timestamp=%d",
		event.VehicleID,
		event.Event,
		event.Location.Latitude,
		event.Location.Longitude,
		event.Timestamp,
	)
}
