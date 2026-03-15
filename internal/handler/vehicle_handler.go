package handler

import (
	"net/http"
	"strconv"

	"github.com/elokanugrah/fleet-management-service/internal/dto"
	"github.com/elokanugrah/fleet-management-service/internal/usecase"
	"github.com/gin-gonic/gin"
)

type VehicleHandler struct {
	usecase usecase.VehicleUsecase
}

func NewVehicleHandler(uc usecase.VehicleUsecase) *VehicleHandler {
	return &VehicleHandler{usecase: uc}
}

func (h *VehicleHandler) RegisterRoutes(r *gin.Engine) {
	vehicles := r.Group("/vehicles")
	{
		vehicles.GET("/:vehicle_id/location", h.GetLastLocation)
		vehicles.GET("/:vehicle_id/history", h.GetHistory)
	}
}

// GetLastLocation godoc
// GET /vehicles/:vehicle_id/location
func (h *VehicleHandler) GetLastLocation(c *gin.Context) {
	vehicleID := c.Param("vehicle_id")

	loc, err := h.usecase.GetLastLocation(c.Request.Context(), vehicleID)
	if err != nil {
		dto.SendError(c, http.StatusInternalServerError, "failed to fetch location")
		return
	}
	if loc == nil {
		dto.SendError(c, http.StatusNotFound, "vehicle not found")
		return
	}

	dto.SendSuccess(c, dto.LocationResponse{
		VehicleID: loc.VehicleID,
		Latitude:  loc.Latitude,
		Longitude: loc.Longitude,
		Timestamp: loc.Timestamp,
	})
}

// GetHistory godoc
// GET /vehicles/:vehicle_id/history?start=...&end=...
func (h *VehicleHandler) GetHistory(c *gin.Context) {
	vehicleID := c.Param("vehicle_id")

	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		dto.SendError(c, http.StatusBadRequest, "start and end query params are required")
		return
	}

	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		dto.SendError(c, http.StatusBadRequest, "invalid start timestamp")
		return
	}

	end, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		dto.SendError(c, http.StatusBadRequest, "invalid end timestamp")
		return
	}

	if start > end {
		dto.SendError(c, http.StatusBadRequest, "start must be before end")
		return
	}

	locations, err := h.usecase.GetHistory(c.Request.Context(), vehicleID, start, end)
	if err != nil {
		dto.SendError(c, http.StatusInternalServerError, "failed to fetch history")
		return
	}

	var data []dto.LocationResponse
	for _, loc := range locations {
		data = append(data, dto.LocationResponse{
			VehicleID: loc.VehicleID,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
			Timestamp: loc.Timestamp,
		})
	}

	if data == nil {
		data = []dto.LocationResponse{}
	}

	dto.SendSuccess(c, data, dto.HistoryResponse{
		VehicleID: vehicleID,
		Total:     len(data),
	})
}
