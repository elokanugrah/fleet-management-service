package usecase

import (
	"context"

	"github.com/elokanugrah/fleet-management-service/internal/domain"
	"github.com/elokanugrah/fleet-management-service/internal/repository"
)

type VehicleUsecase interface {
	GetLastLocation(ctx context.Context, vehicleID string) (*domain.VehicleLocation, error)
	GetHistory(ctx context.Context, vehicleID string, start, end int64) ([]domain.VehicleLocation, error)
}

type vehicleUsecase struct {
	repo repository.VehicleRepository
}

func NewVehicleUsecase(
	repo repository.VehicleRepository,
) VehicleUsecase {
	return &vehicleUsecase{
		repo: repo,
	}
}

func (u *vehicleUsecase) GetLastLocation(ctx context.Context, vehicleID string) (*domain.VehicleLocation, error) {
	return u.repo.GetLastLocation(ctx, vehicleID)
}

func (u *vehicleUsecase) GetHistory(ctx context.Context, vehicleID string, start, end int64) ([]domain.VehicleLocation, error) {
	return u.repo.GetHistory(ctx, vehicleID, start, end)
}
