package mqtt

import (
	"fmt"
	"log"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

func NewClient(broker, clientID string) pahomqtt.Client {
	opts := pahomqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetCleanSession(true)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(30 * time.Second)

	opts.OnConnect = func(c pahomqtt.Client) {
		log.Println("MQTT connected")
	}
	opts.OnConnectionLost = func(c pahomqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	}

	client := pahomqtt.NewClient(opts)
	token := client.Connect()
	if token.WaitTimeout(10*time.Second) && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}

	fmt.Printf("MQTT client connected to %s\n", broker)
	return client
}
