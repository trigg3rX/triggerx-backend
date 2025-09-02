package alchemy

import "fmt"

type Config struct {
	APIKey  string `yaml:"api_key"`
}

var DefaultConfig = Config{
	APIKey:  "",
}

func LoadConfig(apiKey string) (*Config, error) {
	config := DefaultConfig

	if apiKey != "" {
		config.APIKey = apiKey
	}

	return &config, nil
}

// GetEndpoint returns the full endpoint URL for the given method
func (c *Config) GetEndpoint(chainID string) string {
	switch chainID {
	// Testnets
	case "11155111":
		return fmt.Sprintf("https://eth-sepolia.g.alchemy.com/v2/%s", c.APIKey)
	case "11155420":
		return fmt.Sprintf("https://opt-sepolia.g.alchemy.com/v2/%s", c.APIKey)
	case "84532":
		return fmt.Sprintf("https://base-sepolia.g.alchemy.com/v2/%s", c.APIKey)
	case "421614":
		return fmt.Sprintf("https://arb-sepolia.g.alchemy.com/v2/%s", c.APIKey)

	// Mainnets
	case "1":
		return fmt.Sprintf("https://eth-mainnet.g.alchemy.com/v2/%s", c.APIKey)
	case "10":
		return fmt.Sprintf("https://opt-mainnet.g.alchemy.com/v2/%s", c.APIKey)
	case "8453":
		return fmt.Sprintf("https://base-mainnet.g.alchemy.com/v2/%s", c.APIKey)
	case "42161":
		return fmt.Sprintf("https://arb-mainnet.g.alchemy.com/v2/%s", c.APIKey)
	}
	return ""
}
