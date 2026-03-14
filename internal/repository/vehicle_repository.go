package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elokanugrah/fleet-management-service/internal/domain"
)

type VehicleRepository interface {
	Save(ctx context.Context, loc domain.VehicleLocation) error
	GetLastLocation(ctx context.Context, vehicleID string) (*domain.VehicleLocation, error)
	GetHistory(ctx context.Context, vehicleID string, start, end int64) ([]domain.VehicleLocation, error)
}

type vehicleRepository struct {
	db *sql.DB
}

func NewVehicleRepository(db *sql.DB) VehicleRepository {
	return &vehicleRepository{db: db}
}

func (r *vehicleRepository) Save(ctx context.Context, loc domain.VehicleLocation) error {
	query := `
		INSERT INTO vehicle_locations (vehicle_id, latitude, longitude, timestamp)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query,
		loc.VehicleID,
		loc.Latitude,
		loc.Longitude,
		loc.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("repository.Save: %w", err)
	}
	return nil
}

func (r *vehicleRepository) GetLastLocation(ctx context.Context, vehicleID string) (*domain.VehicleLocation, error) {
	query := `
		SELECT vehicle_id, latitude, longitude, timestamp
		FROM vehicle_locations
		WHERE vehicle_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, vehicleID)

	var loc domain.VehicleLocation
	err := row.Scan(&loc.VehicleID, &loc.Latitude, &loc.Longitude, &loc.Timestamp)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetLastLocation: %w", err)
	}
	return &loc, nil
}

func (r *vehicleRepository) GetHistory(ctx context.Context, vehicleID string, start, end int64) ([]domain.VehicleLocation, error) {
	query := `
		SELECT vehicle_id, latitude, longitude, timestamp
		FROM vehicle_locations
		WHERE vehicle_id = $1
		  AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp ASC
	`
	rows, err := r.db.QueryContext(ctx, query, vehicleID, start, end)
	if err != nil {
		return nil, fmt.Errorf("repository.GetHistory: %w", err)
	}
	defer rows.Close()

	var locations []domain.VehicleLocation
	for rows.Next() {
		var loc domain.VehicleLocation
		if err := rows.Scan(&loc.VehicleID, &loc.Latitude, &loc.Longitude, &loc.Timestamp); err != nil {
			return nil, fmt.Errorf("repository.GetHistory scan: %w", err)
		}
		locations = append(locations, loc)
	}
	return locations, nil
}
