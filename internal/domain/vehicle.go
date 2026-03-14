package domain

// VehicleLocation represents location data received from MQTT
type VehicleLocation struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

// Location is used inside GeofenceEvent
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// GeofencePoint is a monitored area in the system
type GeofencePoint struct {
	Name      string
	Latitude  float64
	Longitude float64
	Radius    float64 // in meters
}
