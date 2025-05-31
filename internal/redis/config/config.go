package config

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DevMode bool
}

var cfg Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	cfg = Config{
		DevMode: os.Getenv("DEV_MODE") == "true",
	}

	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	return nil
}
func IsDevMode() bool {
	return cfg.DevMode
}
