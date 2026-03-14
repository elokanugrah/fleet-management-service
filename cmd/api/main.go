package main

import (
	"github.com/elokanugrah/fleet-management-service/config"
	pkgmqtt "github.com/elokanugrah/fleet-management-service/pkg/mqtt"
	"github.com/elokanugrah/fleet-management-service/pkg/postgres"
	"github.com/elokanugrah/fleet-management-service/pkg/rabbitmq"
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
}
