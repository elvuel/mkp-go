package main

import (
	"context"
	"fmt"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	mkpgo "github.com/elvuel/mkp-go"
	servercfg "github.com/elvuel/mkp-go/cmd/server/config"
	"github.com/elvuel/mkp-go/cmd/server/controllers"
)

func main() {
	configPath := flag.String("c", "", "path to server config JSON")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("mkp-server %s\n", Version)
		return
	}

	appCfg, loadedPath, err := servercfg.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if loadedPath != "" {
		log.Printf("loaded config file: %s", loadedPath)
	}

	sfportName := appCfg.SFPort
	if sfportName == "" {
		detectedPort, err := mkpgo.CheckSFSerialPort()
		if err != nil {
			log.Fatalf("SF_PORT is not set and auto-detect failed: %v", err)
		}
		sfportName = detectedPort
		log.Printf("SF_PORT is not set, auto-detected serial port: %s", sfportName)
	}

	db, err := initDatabase(appCfg.Database)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	api, err := controllers.NewAPI(sfportName, appCfg.Auth, db)
	if err != nil {
		log.Fatalf("failed to initialize api controller: %v", err)
	}
	defer api.Close()

	if appCfg.Auth.JWTSecret == "mkp-go-dev-secret" {
		log.Println("JWT_SECRET is not set, using default development secret")
	}

	gin.SetMode(appCfg.Mode)
	log.Printf("server running mode: %s", appCfg.Mode)
	if appCfg.Mode == "debug" && appCfg.Auth.MuteInDebug {
		log.Println("JWT auth middleware is muted in debug mode")
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		MaxAge:       12 * time.Hour,
	}))
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true, "status": "up"})
	})
	api.RegisterRoutes(router)

	httpServer := &http.Server{
		Addr:         ":" + appCfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("gin server listening on :%s (SF_PORT=%s)", appCfg.Port, sfportName)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}
}
