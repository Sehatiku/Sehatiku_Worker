package config

import (
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func NewViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName(".env")
	v.SetConfigType("env")

	dir := "."
	for i := 0; i < 5; i++ {
		v.AddConfigPath(dir)
		dir = filepath.Join(dir, "..")
	}
	_ = v.ReadInConfig()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return v
}
