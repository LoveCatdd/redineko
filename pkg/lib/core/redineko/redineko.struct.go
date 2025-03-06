package redineko

import (
	"github.com/LoveCatdd/util/pkg/lib/core/viper"
)

type RedisConfig struct {
	Redis struct {
		Enable      bool   `mapstructure:"enable"`
		MaxIdle     int    `mapstructure:"maxIdle"`
		MaxActive   int    `mapstructure:"maxActive"`
		IdleTimeout int32  `mapstructure:"idleTimeout"`
		Ip          string `mapstructure:"ip"`
		Port        string `mapstructure:"port"`
		Password    string `mapstructure:"password"`
	} `mapstructure:"redis"`
}

func (RedisConfig) FileType() string {
	return viper.VIPER_YAML
}

var RediConf = new(RedisConfig)
