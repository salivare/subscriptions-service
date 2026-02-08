package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config is the main application configuration structure.
type Config struct {
	Env           string         `yaml:"env" env-default:"local"`
	HTTPServer    HTTPConfig     `yaml:"http_server"`
	Postgres      PostgresConfig `yaml:"postgres"`
	SwaggerServer SwaggerConfig  `yaml:"swagger_server"`
}

// HTTPConfig defines the parameters for the underlying http.Server.
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

type PostgresConfig struct {
	Host            string      `yaml:"host" env-default:"localhost"`
	Port            int         `yaml:"port" env-default:"5432"`
	User            string      `yaml:"user" env-default:"postgres"`
	Password        string      `yaml:"password" env-default:""`
	DBName          string      `yaml:"dbname" env-default:"postgres"`
	SSLMode         string      `yaml:"sslmode" env-default:"disable"`
	MigrationsPath  string      `yaml:"migrations_path" env-default:"./migrations"`
	MigrationsTable string      `yaml:"migrations_table" env-default:"schema_migrations"`
	Retry           RetryConfig `yaml:"retry"`
}

type RetryConfig struct {
	Attempts     int           `yaml:"attempts" env-default:"10"`
	InitialDelay time.Duration `yaml:"initial_delay" env-default:"1s"`
	MaxDelay     time.Duration `yaml:"max_delay" env-default:"10s"`
	Step         time.Duration `yaml:"step" env-default:"2s"`
}

type SwaggerConfig struct {
	JSONPath string `yaml:"json_path" env-default:"./swagger.json"`
	UIPath   string `yaml:"ui_path" env-default:"./swaggerui"`
}

// MustLoad reads the configuration from the path provided via flags or environment variables.
// It panics if the configuration cannot be loaded.
func MustLoad() *Config {
	path := fetchConfigPath()

	if path == "" {
		panic("config file path is empty")
	}

	return MustLoadByPath(path)
}

// MustLoadByPath reads the configuration from a specific file path.
// It panics if the file is missing or invalid.
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
