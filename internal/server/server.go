package server

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"qtpi/internal/sensor"
	"qtpi/internal/ws"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func Server() {
	//Parse command line flags
	host := flag.String("host", "localhost", "Server host")
	port := flag.Int("port", 8080, "Server port")
	flag.Parse()

	gin.SetMode(gin.ReleaseMode)

	apiRouter := gin.Default()
	api := apiRouter.Group("/api")
	sensor.AddSensorRoutes(api)

	sh := ws.NewHub[*sensor.SensorClient]()
	go sh.Run()

	sensor.AddSensorWs(api, sh)

	serverPath := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("Server started at http://%s ...\n", serverPath)

	server := &http.Server{
		Addr:         serverPath,
		Handler:      apiRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// start server
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln(err)
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
	log.Println("Server exiting")
}
