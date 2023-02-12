// Package config all the configurable variables within notification-service.
//
// Configs can be loaded from an external .toml file as well as via environment variables.
// Defaults are set within this package.
//
// When both environment variables and config file variables are loaded, the order of precedence is:
//
// Environment > Config file > Defaultdsp
package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Configuration is top-level and represents all configuration settings
type Configuration struct {
	// Application
	App AppConfig
	// Log
	Log LogConfig
	// Service
	// Service ServiceConfig
	//
	Services []ServiceConfig
	// Kafka
	Kafka KafkaConfig
}

// AppConfig represents the application config
type AppConfig struct {
	// Port for http server
	Port int `mapstructure:"Port"`
	// Enable pprof
	EnablePprof bool `mapstructure:"EnablePprof"`
}

type KafkaConfig struct {
	UseCredentials bool     `mapstructure:"UseCredentials"`
	Username       string   `mapstructure:"Username"`
	Password       string   `mapstructure:"Password"`
	Brokers        []string `mapstructure:"Brokers"`
}

type ServiceConfig struct {
	Name        string        `mapstructure:"Name"`
	MaxRequests int           `mapstructure:"MaxRequests"`
	TenantID    string        `mapstructure:"TenantId"`
	Retry       int           `mapstructure:"Retry"`
	RetryDelay  time.Duration `mapstructure:"RetryDelay"`
	Timeout     time.Duration `mapstructure:"Timeout"`
	GroupID     string        `mapstructure:"GroupID"`
	Topic       string        `mapstructure:"Topic"`
	Error       string        `mapstructure:"Error"`
}

// LogConfig represents the log configuration
type LogConfig struct {
	// Log level
	Level int `mapstructure:"Level"`
	// Verbose mode will allow to append logging middlewares
	Verbose bool `mapstructure:"Verbose"`
}

var config *Configuration
var once sync.Once

// for unit test
var once2 sync.Once

// GetConfig initializes and return the top-level Configuration struct
func GetConfig() *Configuration {
	if config == nil {
		once.Do(func() {
			setup()
		})
	}

	return config
}

func setup() {
	config = &Configuration{}

	bindEnv()

	// For unit test, bindflag can't be called twice
	once2.Do(func() {
		bindFlag()
	})

	bindFile()

	err := viper.ReadInConfig()
	if err != nil {
		log.Error().Err(err).Msg("error when reading the config file")
	}

	err = viper.Unmarshal(config)
	if err != nil {
		log.Fatal().Err(err).Msg("unmarshal config")
	}

	err = validate()
	if err != nil {
		log.Fatal().Err(err).Msg("validate config")
		return
	}

	setServiceVariables()
}

func validate() error {
	for i := range config.Services {
		service := config.Services[i]

		if service.Name == "" {
			return errors.Wrap(ErrRequiredParameter, "name")
		}

		if service.TenantID == "" {
			return errors.Wrap(ErrRequiredParameter, "tenantId")
		}
	}

	return nil
}

func setServiceVariables() {
	for i := range config.Services {
		service := &config.Services[i]

		if service.MaxRequests < 1 {
			service.MaxRequests = 1
		}

		service.Topic = fmt.Sprintf("%s-%s", service.TenantID, service.Name)
		service.Error = fmt.Sprintf("%s-%s-error", service.TenantID, service.Name)
	}
}

func bindEnv() {
	viper.SetEnvPrefix("NOTIFICATION_SERVICE")

	viper.BindEnv("Log.Level", "NOTIFICATION_SERVICE_LOG_LEVEL")

	viper.AutomaticEnv()
}

func bindFile() {
	filePath := viper.GetString("CONFIG_PATH")
	if _, err := os.Stat(filePath); err == nil {
		ftype := filepath.Ext(filePath)
		if len(ftype) > 1 {
			ftype = ftype[1:]
		}
		fname := filepath.Base(filePath)
		fname = fname[0 : len(fname)-(len(ftype)+1)]
		fpath := filepath.Dir(filePath)
		viper.SetConfigName(fname)
		viper.SetConfigType(ftype)
		viper.AddConfigPath(fpath)
	} else {
		viper.SetConfigName("notification-service")
		viper.SetConfigType("toml")
		viper.AddConfigPath("./")
		viper.AddConfigPath("/etc/notification-service/conf.d/")
	}
}

func bindFlag() {
	pflag.IntP("APP.PORT", "p", 11000, "port of the api")
	pflag.StringP("CONFIG_PATH", "c", "/etc/notification-service/conf.d/notification-service.toml", "location of the config file")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		log.Printf("unexpected error while Binding Flags %s", err)
	}
}
