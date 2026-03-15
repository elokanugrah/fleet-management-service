package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

type LocationPayload struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

func main() {
	broker := getEnv("MQTT_BROKER", "tcp://localhost:1883")
	vehicleID := getEnv("VEHICLE_ID", "B1234XYZ")

	// Bus stop point (geofence center)
	halteLat := -6.2088
	halteLng := 106.8456

	// Bus route simulation: from point A passing through bus stop to point B
	// Point A: ~600m before bus stop (west)
	// Point B: ~600m after bus stop (east)
	startLat := halteLat - 0.001
	startLng := halteLng - 0.005
	endLat := halteLat + 0.001
	endLng := halteLng + 0.005

	// Total ticks in one travel cycle A -> B
	// 90 ticks = 180 seconds per cycle
	// Distance per tick: ~600m / 90 ticks approx 6.7m per tick
	// Bus within 50m radius approx 100m / 6.7m approx ~15 ticks (~30 seconds)
	const totalTicks = 90

	opts := pahomqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("mock-gps-publisher-" + vehicleID)
	opts.SetCleanSession(true)
	opts.OnConnect = func(c pahomqtt.Client) {
		log.Println("Mock GPS publisher connected to MQTT")
	}

	client := pahomqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect: %v", token.Error())
	}
	defer client.Disconnect(250)

	topic := fmt.Sprintf("/fleet/vehicle/%s/location", vehicleID)
	log.Printf("Publishing to topic: %s every 2 seconds", topic)
	log.Printf("Route: (%.4f, %.4f) -> Bus Stop (%.4f, %.4f) -> (%.4f, %.4f)",
		startLat, startLng, halteLat, halteLng, endLat, endLng)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	i := 0
	for range ticker.C {
		// Travel progress: 0.0 (point A) -> 1.0 (point B)
		progress := float64(i%totalTicks) / float64(totalTicks)

		// Linear interpolation from A to B (bus moves gradually)
		lat := startLat + (endLat-startLat)*progress
		lng := startLng + (endLng-startLng)*progress

		// Add small noise so it's not too straight (~5m)
		lat += (rand.Float64()*0.0001 - 0.00005)
		lng += (rand.Float64()*0.0001 - 0.00005)

		// Calculate distance to bus stop for log label
		dist := haversineDistance(lat, lng, halteLat, halteLng)
		status := "outside geofence "
		if dist <= 50 {
			status = "INSIDE geofence"
		}

		payload := LocationPayload{
			VehicleID: vehicleID,
			Latitude:  lat,
			Longitude: lng,
			Timestamp: time.Now().Unix(),
		}

		body, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal payload: %v", err)
			continue
		}

		token := client.Publish(topic, 1, false, body)
		token.Wait()
		if token.Error() != nil {
			log.Printf("Failed to publish: %v", token.Error())
			continue
		}

		log.Printf("[tick %d | %s | %.0fm dari halte] vehicle=%s lat=%.6f lng=%.6f",
			i+1, status, dist, vehicleID, lat, lng)
		i++
	}
}

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
