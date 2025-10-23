package config

import (
	"fmt"
	"strconv"

	"github.com/spf13/viper"
)

var AppConfig Config

type Config struct {
	Log              LogConfig
	Server           ServerConfig
	Environment      string
	ProfilingEnabled bool
	Kafka            KafkaConfig
	Clients          Clients
}

type LogConfig struct {
	Level string
}

type ServerConfig struct {
	Host string
	Port int
}

type SvcConfig struct {
	Host string
	Port string
}

type Clients struct {
	Entities Entities
}

type Entities struct {
	Svc         SvcConfig
	TimeoutInMS EntitiesTimeouts
}

type EntitiesTimeouts struct {
	GetUser int `mapstructure:"getUser"`
}

func InitDefaultConfig() *Config {
	return InitConfig("application")
}

func InitConfig(configname string) *Config {
	viper.AutomaticEnv()
	viper.SetConfigName(configname)
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("rapido")
	viper.AddConfigPath("config")
	viper.AddConfigPath("../config/")
	viper.AddConfigPath("../../config/")
	viper.AddConfigPath("../../../config/")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("Cannot read config file config/application.yaml")
		panic(err)
	}

	err := viper.UnmarshalExact(&AppConfig)
	if err != nil {
		panic(fmt.Errorf("fatal error unable to Unmarshal config file: %s", err))
	}

	return &AppConfig
}

func GetConfig() *Config {
	return &AppConfig
}

func (c *Config) IsProductionEnv() bool {
	return c.Environment == "production"
}

func (c *Config) ListenAddress() string {
	return ":" + strconv.Itoa(c.Server.Port)
}
