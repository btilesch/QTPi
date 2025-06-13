package sensor

import (
	"net/http"
	"qtpi/internal/ws"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type SensorData struct {
	Id   string  `json:"id"`
	Temp float64 `json:"temp"`
	Time string  `json:"time"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func sensorLatestHandler(c *gin.Context) {
	data := SensorData{
		Time: time.Now().String(),
		Temp: 22.56,
	}

	c.JSON(http.StatusOK, data)
}

func sensorServeWs(c *gin.Context, h *ws.Hub[*SensorClient]) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to websocket"})
		return
	}

	NewSensorClient(conn, h)
}
