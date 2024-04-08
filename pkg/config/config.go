package config

import (
	"os"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type (
	Server struct {
		Addr string `yaml:"addr"`
		Port string `yaml:"port"`
	}

	Database struct {
		Name     string `yaml:"name"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Ssl      string `yaml:"ssl"`
	}

	Proxy struct {
		Addr string `yaml:"addr"`
		Port string `yaml:"port"`
	}

	Logger struct {
		Level string `yaml:"addr"`
	}
)

type Config struct {
	Server   Server   `yaml:"server"`
	Database Database `yaml:"database"`
	Proxy    Proxy    `yaml:"proxy"`
	Logger   Logger   `yaml:"logger"`
}

func GetConfig(cfgPath string) (Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(cfgPath)

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(&cfg, viper.DecodeHook(StringExpandEnv())); err != nil {
		return Config{}, err
	}

	return *cfg, nil
}

func StringExpandEnv() mapstructure.DecodeHookFuncKind {
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{}) (interface{}, error) {
		if f != reflect.String || t != reflect.String {
			return data, nil
		}
		return os.ExpandEnv(data.(string)), nil
	}
}
