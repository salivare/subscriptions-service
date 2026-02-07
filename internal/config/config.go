package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string     `yaml:"env" env-default:"local"`
	HTTPServer HTTPConfig `yaml:"http_server"`
}

type HTTPConfig struct {
	Host              string        `yaml:"host" env-default:"localhost"`
	Port              int           `yaml:"port" env-default:"8080"`
	Timeout           time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout       time.Duration `yaml:"idle_timeout" env-default:"60s"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout" env-default:"10s"`
	ReadTimeout       time.Duration `yaml:"read_timeout" env-default:"5s"`
	WriteTimeout      time.Duration `yaml:"write_timeout" env-default:"10s"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" env-default:"2s"`
	MaxHeaderBytes    int           `yaml:"max_header_bytes" env-default:"1048576"`
}

func MustLoad() *Config {
	path := fetchConfigPath()

	if path == "" {
		panic("config file path is empty")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
