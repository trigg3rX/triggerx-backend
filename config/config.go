package config

import (
	"github.com/spf13/viper"
    "github.com/trigg3rX/eigensdk-go/types"
)

func LoadConfig() (*types.QuorumNum, error) {
    viper.SetConfigFile("triggerx_operator.yaml")
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }
    var config types.QuorumNum
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }
    return &config, nil
}
