package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/elokanugrah/fleet-management-service/internal/domain"
	"github.com/elokanugrah/fleet-management-service/internal/usecase"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

const locationTopic = "/fleet/vehicle/+/location"

type LocationSubscriber struct {
	client  pahomqtt.Client
	usecase usecase.VehicleUsecase
}

func NewLocationSubscriber(client pahomqtt.Client, uc usecase.VehicleUsecase) *LocationSubscriber {
	return &LocationSubscriber{
		client:  client,
		usecase: uc,
	}
}

// Start subscribes to the MQTT topic and starts listening
func (s *LocationSubscriber) Start() {
	token := s.client.Subscribe(locationTopic, 1, s.handleMessage)
	token.Wait()
	if token.Error() != nil {
		log.Fatalf("Failed to subscribe to MQTT topic: %v", token.Error())
	}
	log.Printf("MQTT subscribed to topic: %s", locationTopic)
}

func (s *LocationSubscriber) handleMessage(client pahomqtt.Client, msg pahomqtt.Message) {
	log.Printf("[MQTT] Received message on topic: %s", msg.Topic())

	var loc domain.VehicleLocation
	if err := json.Unmarshal(msg.Payload(), &loc); err != nil {
		log.Printf("[MQTT] Failed to parse payload: %v | raw: %s", err, string(msg.Payload()))
		return
	}

	// Validate required fields
	if err := validateLocation(loc); err != nil {
		log.Printf("[MQTT] Invalid payload: %v", err)
		return
	}

	ctx := context.Background()
	if err := s.usecase.ProcessLocation(ctx, loc); err != nil {
		log.Printf("[MQTT] Failed to process location: %v", err)
		return
	}

	log.Printf("[MQTT] Processed location for vehicle %s: (%.6f, %.6f)",
		loc.VehicleID, loc.Latitude, loc.Longitude)
}

func validateLocation(loc domain.VehicleLocation) error {
	if loc.VehicleID == "" {
		return fmt.Errorf("vehicle_id is required")
	}
	if loc.Latitude == 0 && loc.Longitude == 0 {
		return fmt.Errorf("latitude and longitude cannot both be zero")
	}
	if loc.Timestamp == 0 {
		return fmt.Errorf("timestamp is required")
	}
	return nil
}
