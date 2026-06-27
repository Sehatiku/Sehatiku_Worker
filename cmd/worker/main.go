package main

import (
	"os"
	"os/signal"
	"sehatiku-notification-worker/internal/config"
	"syscall"
)

func main() {
	cfg := config.NewViper()
	log := config.NewLogger(cfg)
	db := config.ConnectDB(cfg, log)
	wa := config.SetUpWhatsApp(cfg, log, db)

	scheduler := config.BootStrap(&config.BootStrapConfig{
		DB:       db,
		Log:      log,
		Config:   cfg,
		WhatsApp: wa,
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down worker")
	scheduler.Stop()
	wa.Client.Disconnect()
	log.Info("worker stopped")
}
