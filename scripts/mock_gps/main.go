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
	// Point A: ~500m before bus stop
	// Point B: ~500m after bus stop
	startLat := halteLat - 0.0045
	startLng := halteLng - 0.0045
	endLat := halteLat + 0.0045
	endLng := halteLng + 0.0045

	// Speed in degrees per tick
	const (
		normalStep = 0.00027 // ~30m per tick (normal speed)
		slowStep   = 0.00005 // ~5m per tick  (slow speed approaching bus stop)
		slowRadius = 150.0   // start slowing down at 150m radius from bus stop
	)

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

	// Initial bus position at point A
	currentLat := startLat
	currentLng := startLng
	direction := 1.0

	const stopTicks = 3
	stopCounter := 0
	isStopped := false
	hasStoppedThisPass := false // prevent repeated stops in one pass

	for range ticker.C {
		dist := haversineDistance(currentLat, currentLng, halteLat, halteLng)

		// Trigger stop only once per pass, when first entering 20m radius
		if !isStopped && !hasStoppedThisPass && dist <= 20 {
			isStopped = true
			stopCounter = stopTicks
			log.Printf("--- Bus stopping at bus stop (%.0fm from center) ---", dist)
		}

		if isStopped {
			stopCounter--
			if stopCounter <= 0 {
				isStopped = false
				hasStoppedThisPass = true
				log.Println("--- Bus departing from bus stop ---")
			}
		} else {
			step := normalStep
			if dist <= slowRadius {
				ratio := dist / slowRadius
				step = slowStep + (normalStep-slowStep)*ratio
			}

			currentLat += direction * step
			currentLng += direction * step

			// Small noise ~2m
			currentLat += rand.Float64()*0.00004 - 0.00002
			currentLng += rand.Float64()*0.00004 - 0.00002

			// Reverse direction + reset flag when reaching route end
			if currentLng >= endLng {
				currentLat = endLat
				currentLng = endLng
				direction = -1.0
				hasStoppedThisPass = false
				log.Println("--- Bus reached point B, turning around ---")
			} else if currentLng <= startLng {
				currentLat = startLat
				currentLng = startLng
				direction = 1.0
				hasStoppedThisPass = false
				log.Println("--- Bus returned to point A, moving forward again ---")
			}
		}

		dist = haversineDistance(currentLat, currentLng, halteLat, halteLng)
		status := "outside geofence "
		if dist <= 50 {
			status = "INSIDE geofence"
		}

		speedKmh := 0.0
		if !isStopped {
			step := normalStep
			if dist <= slowRadius {
				ratio := dist / slowRadius
				step = slowStep + (normalStep-slowStep)*ratio
			}
			speedKmh = step * 111000 / 2 * 3.6
		}

		payload := LocationPayload{
			VehicleID: vehicleID,
			Latitude:  currentLat,
			Longitude: currentLng,
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

		log.Printf("[%s | %.0fm from bus stop | %.1f km/h] lat=%.6f lng=%.6f",
			status, dist, speedKmh, currentLat, currentLng)
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
