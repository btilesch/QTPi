package sensor

import (
	"qtpi/internal/ws"

	"github.com/gin-gonic/gin"
)

func AddSensorRoutes(rg *gin.RouterGroup) {
	sensor := rg.Group("/sensor")
	sensor.GET("/current", sensorLatestHandler)
}

func AddSensorWs(rg *gin.RouterGroup, h *ws.Hub[*SensorClient]) {
	sensor := rg.Group("/sensor")
	sensor.GET("/ws", func(c *gin.Context) {
		sensorServeWs(c, h)
	})
}
