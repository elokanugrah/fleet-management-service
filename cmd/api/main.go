package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elokanugrah/fleet-management-service/config"
	"github.com/elokanugrah/fleet-management-service/internal/domain"
	"github.com/elokanugrah/fleet-management-service/internal/handler"
	"github.com/elokanugrah/fleet-management-service/internal/publisher"
	"github.com/elokanugrah/fleet-management-service/internal/repository"
	"github.com/elokanugrah/fleet-management-service/internal/subscriber"
	"github.com/elokanugrah/fleet-management-service/internal/usecase"
	pkgmqtt "github.com/elokanugrah/fleet-management-service/pkg/mqtt"
	"github.com/elokanugrah/fleet-management-service/pkg/postgres"
	"github.com/elokanugrah/fleet-management-service/pkg/rabbitmq"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	cfg := config.Load()

	// Init dependencies
	db := postgres.NewConnection(cfg.DBConnectionString())
	defer db.Close()

	mqttClient := pkgmqtt.NewClient(cfg.MQTTBroker, cfg.MQTTClientID)
	defer mqttClient.Disconnect(250)

	rmqConn := rabbitmq.NewConnection(cfg.RabbitMQURL)
	defer rmqConn.Close()

	// Setup Layers
	vehicleRepo := repository.NewVehicleRepository(db)
	geofencePub := publisher.NewGeofencePublisher(rmqConn)

	// Define geofence points
	geofences := []domain.GeofencePoint{
		{
			Name:      "Manggarai",
			Latitude:  cfg.GeofenceLat,
			Longitude: cfg.GeofenceLng,
			Radius:    cfg.GeofenceRadius,
		},
	}

	vehicleUC := usecase.NewVehicleUsecase(vehicleRepo, geofencePub, geofences)

	// Start MQTT subscriber
	locationSub := subscriber.NewLocationSubscriber(mqttClient, vehicleUC)
	locationSub.Start()

	// Setup HTTP server
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	vehicleHandler := handler.NewVehicleHandler(vehicleUC)
	vehicleHandler.RegisterRoutes(router)

	requestTimeout := time.Duration(cfg.RequestTimeout) * time.Second
	serverAddr := fmt.Sprintf(":%s", cfg.Port)

	srv := &http.Server{
		Addr:              serverAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second, // Protect against Slowloris attacks
		WriteTimeout:      requestTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("API server starting on :%s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Server stopped")

}
