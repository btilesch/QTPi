package sensor

import (
	"encoding/json"
	"home-monitor/internal/ws"
	"log"

	"github.com/gorilla/websocket"
)

type SensorClient struct {
	Client        *ws.Client
	Subscriptions map[string]bool
}

func NewSensorClient(conn *websocket.Conn, hub *ws.Hub) *SensorClient {
	base := ws.NewClient(hub, conn)
	sc := &SensorClient{
		Client:        base,
		Subscriptions: make(map[string]bool),
	}
	hub.Register <- base

	go base.ReadPump()
	go base.WritePump()
	go sc.handleMessages()

	return sc
}

func (sc *SensorClient) handleMessages() {
	for msg := range sc.Client.Receive {
		var payload struct {
			Subscribe []string `json:"subscribe"`
		}
		if err := json.Unmarshal(msg, &payload); err != nil {
			continue
		}
		for _, s := range payload.Subscribe {
			sc.Subscriptions[s] = true
			log.Printf("Client %v subscribed to %v", sc.Client.Id, s)
		}
	}
}
