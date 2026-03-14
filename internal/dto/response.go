package dto

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// LocationResponse is the API response for last known location
type LocationResponse struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

// HistoryResponse is the API response for location history
type HistoryResponse struct {
	VehicleID string `json:"vehicle_id"`
	Total     int    `json:"total"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    interface{} `json:"meta,omitempty"`
}

// ErrorResponse is the standard error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SendSuccess sends a standardized success response
func SendSuccess(c *gin.Context, data interface{}, meta ...interface{}) {
	response := SuccessResponse{
		Success: true,
		Data:    data,
	}

	if len(meta) > 0 {
		response.Meta = meta[0]
	}

	c.JSON(http.StatusOK, response)
}

// SendError sends a standardized error response
func SendError(c *gin.Context, status int, message string) {
	c.JSON(status, ErrorResponse{
		Success: false,
		Message: message,
	})
}
