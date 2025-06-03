package sensor

import (
	"encoding/json"
	"home-monitor/internal/ws"
	"log"

	"github.com/gorilla/websocket"
)

var sensorClients = make(map[string]*SensorClient)

type SensorClient struct {
	*ws.BaseClient
	Subscriptions map[string]bool
}

func NewSensorClient(conn *websocket.Conn, hub *ws.Hub[*SensorClient]) *SensorClient {
	sc := &SensorClient{
		Subscriptions: make(map[string]bool),
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

func (sc *SensorClient) handleUnsubscribe(h *ws.Hub[*SensorClient]) {
	for msg := range sc.Unregister() {
		h.Unregister <- sc
	}
}
