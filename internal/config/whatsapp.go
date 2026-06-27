package config

import (
	"context"
	"sehatiku-notification-worker/internal/gateway/whatsapp"

	"github.com/spf13/viper"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func SetUpWhatsApp(_ *viper.Viper, log *zap.Logger, db *gorm.DB) *whatsapp.WhatsAppGateway {
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get sql.DB for whatsapp sqlstore", zap.Error(err))
	}

	waLogger := waLog.Stdout("whatsmeow", "WARN", true)
	container := sqlstore.NewWithDB(sqlDB, "postgres", waLogger)
	if err := container.Upgrade(context.Background()); err != nil {
		log.Fatal("failed to upgrade whatsapp schema", zap.Error(err))
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		log.Fatal("failed to get whatsapp device", zap.Error(err))
	}

	client := whatsmeow.NewClient(deviceStore, waLogger)
	client.EnableAutoReconnect = true

	if client.Store.ID == nil {
		log.Warn("whatsapp client not paired — run 'make setup-wa' to pair")
	} else {
		if err := client.Connect(); err != nil {
			log.Warn("whatsapp connect failed", zap.Error(err))
		} else {
			log.Info("whatsapp client connected")
		}
	}

	return whatsapp.New(client, log)
}
