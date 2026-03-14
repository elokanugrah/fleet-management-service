package usecase

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/elokanugrah/fleet-management-service/internal/domain"
	"github.com/elokanugrah/fleet-management-service/internal/publisher"
	"github.com/elokanugrah/fleet-management-service/internal/repository"
)

type VehicleUsecase interface {
	GetLastLocation(ctx context.Context, vehicleID string) (*domain.VehicleLocation, error)
	GetHistory(ctx context.Context, vehicleID string, start, end int64) ([]domain.VehicleLocation, error)
}

type vehicleUsecase struct {
	repo      repository.VehicleRepository
	publisher publisher.GeofencePublisher
	geofences []domain.GeofencePoint
}

func NewVehicleUsecase(
	repo repository.VehicleRepository,
	pub publisher.GeofencePublisher,
	geofences []domain.GeofencePoint,
) VehicleUsecase {
	return &vehicleUsecase{
		repo:      repo,
		publisher: pub,
		geofences: geofences,
	}
}

func (u *vehicleUsecase) ProcessLocation(ctx context.Context, loc domain.VehicleLocation) error {
	// Save to database
	if err := u.repo.Save(ctx, loc); err != nil {
		return fmt.Errorf("usecase.ProcessLocation save: %w", err)
	}

	// Check geofence for every monitored point
	for _, gf := range u.geofences {
		dist := haversineDistance(loc.Latitude, loc.Longitude, gf.Latitude, gf.Longitude)
		if dist <= gf.Radius {
			log.Printf("[Geofence] Vehicle %s entered zone '%s' (%.2fm away)", loc.VehicleID, gf.Name, dist)

			event := domain.GeofenceEvent{
				VehicleID: loc.VehicleID,
				Event:     "geofence_entry",
				Location: domain.Location{
					Latitude:  loc.Latitude,
					Longitude: loc.Longitude,
				},
				Timestamp: time.Now().Unix(),
			}

			if err := u.publisher.Publish(ctx, event); err != nil {
				log.Printf("[Geofence] Failed to publish event: %v", err)
			}
		}
	}

	return nil
}

func (u *vehicleUsecase) GetLastLocation(ctx context.Context, vehicleID string) (*domain.VehicleLocation, error) {
	return u.repo.GetLastLocation(ctx, vehicleID)
}

func (u *vehicleUsecase) GetHistory(ctx context.Context, vehicleID string, start, end int64) ([]domain.VehicleLocation, error) {
	return u.repo.GetHistory(ctx, vehicleID, start, end)
}

func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meters

	lat1Rad := toRad(lat1)
	lat2Rad := toRad(lat2)
	deltaLat := toRad(lat2 - lat1)
	deltaLon := toRad(lon2 - lon1)

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func toRad(deg float64) float64 {
	return deg * math.Pi / 180
}
