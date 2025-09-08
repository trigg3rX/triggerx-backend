package ipfs

import (
	"fmt"
	"strings"
)

type Config struct {
	PinataHost    string
	PinataJWT     string
	PinataBaseURL string
}

func NewConfig(pinataHost string, pinataJWT string) *Config {
	return &Config{
		PinataHost:    pinataHost,
		PinataJWT:     pinataJWT,
		PinataBaseURL: "https://uploads.pinata.cloud/v3/files",
	}
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.PinataHost) == "" {
		return fmt.Errorf("PinataHost is required")
	}
	if strings.TrimSpace(c.PinataJWT) == "" {
		return fmt.Errorf("PinataJWT is required")
	}
	return nil
}
