package sensor

import (
	"encoding/json"
	"log"
	"qtpi/internal/ws"

	"github.com/gorilla/websocket"
)

type SensorClient struct {
	*ws.BaseClient
	Subscriptions map[string]bool
}

func NewSensorClient(conn *websocket.Conn, hub *ws.Hub[*SensorClient]) *SensorClient {
	sc := &SensorClient{
		BaseClient:    ws.NewBaseClient(conn),
		Subscriptions: make(map[string]bool),
	}

	sc.OnUnregister = func() {
		hub.Unregister <- sc
	}

	hub.Register <- sc

	go sc.ReadPump()
	go sc.WritePump()
	go sc.handleMessages()

	return sc
}

func (sc *SensorClient) handleMessages() {
	for msg := range sc.Receive() {
		var payload struct {
			Subscribe []string `json:"subscribe"`
		}
		if err := json.Unmarshal(msg, &payload); err != nil {
			continue
		}
		for _, s := range payload.Subscribe {
			sc.Subscriptions[s] = true
			log.Printf("Client %v subscribed to %v", sc.Id(), s)
		}
	}
}
