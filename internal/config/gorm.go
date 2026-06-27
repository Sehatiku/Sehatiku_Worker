package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func ConnectDB(v *viper.Viper, log *zap.Logger) *gorm.DB {
	dsn := v.GetString("DATABASE_URL")
	if !strings.Contains(dsn, "sslmode=") {
		if strings.Contains(dsn, "?") {
			dsn += "&sslmode=require"
		} else {
			dsn += "?sslmode=require"
		}
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get sql.DB", zap.Error(err))
	}
	sqlDB.SetMaxIdleConns(3)
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	log.Info("database connected")
	return db
}
