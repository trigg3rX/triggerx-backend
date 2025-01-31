package utils

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func HandleHomeDirPath(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

func GetConfig(configPath string) (config types.NodeConfig, err error) {
	absPath := HandleHomeDirPath(configPath)
	yamlFile, err := os.ReadFile(absPath)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(yamlFile, &config)
	return config, err
}

