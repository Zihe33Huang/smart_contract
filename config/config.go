package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	config     *viper.Viper
	configPath = "./conf/"
)

func Config() *viper.Viper {
	if config == nil {
		Init()
	}
	return config
}

func Init(paths ...string) {
	if len(paths) != 0 {
		configPath = paths[0] + "/"
	}
	config = initConfig(configPath)
	fmt.Println("Loading configuration logics...")
	go dynamicConfig()
}

func initConfig(path string) *viper.Viper {
	GlobalConfig := viper.New()
	GlobalConfig.AddConfigPath(path)
	GlobalConfig.SetConfigName("config")
	GlobalConfig.SetConfigType("toml")
	err := GlobalConfig.ReadInConfig()
	if err != nil {
		fmt.Printf("Failed to get the configuration.")
	}
	return GlobalConfig
}

func dynamicConfig() {
	config.WatchConfig()
	config.OnConfigChange(func(event fsnotify.Event) {
		fmt.Printf("Detect config change: %s \n", event.String())
	})
}
