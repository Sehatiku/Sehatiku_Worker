package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sehatiku-notification-worker/internal/config"

	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	cfg := config.NewViper()
	log := config.NewLogger(cfg)
	db := config.ConnectDB(cfg, log)

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get sql.DB: %v\n", err)
		os.Exit(1)
	}

	waLogger := waLog.Stdout("wa-setup", "DEBUG", true)
	container := sqlstore.NewWithDB(sqlDB, "postgres", waLogger)
	_ = container.Upgrade(context.Background())

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get device: %v\n", err)
		os.Exit(1)
	}

	client := whatsmeow.NewClient(deviceStore, waLogger)

	if client.Store.ID != nil {
		fmt.Println("WhatsApp already paired. No action needed.")
		return
	}

	qrChan, _ := client.GetQRChannel(context.Background())
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "connect error: %v\n", err)
		os.Exit(1)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for {
		select {
		case evt := <-qrChan:
			switch evt.Event {
			case whatsmeow.QRChannelEventCode:
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("Scan QR code above with WhatsApp to pair.")
			case "success":
				fmt.Println("Pairing successful!")
				client.Disconnect()
				return
			case "timeout":
				fmt.Println("QR timed out. Restart to try again.")
				return
			}
		case <-c:
			client.Disconnect()
			return
		}
	}
}
