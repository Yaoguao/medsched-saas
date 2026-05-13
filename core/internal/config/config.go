package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env           string        `yaml:"env"`
	HTTPServer    HTTPServer    `yaml:"http_server"`
	StorageConfig StorageConfig `yaml:"storage"`
	Auth          Auth          `yaml:"auth"`
	ML            ML            `yaml:"ml"`
}

type HTTPServer struct {
	Address     string        `yaml:"address"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type StorageConfig struct {
	Postgres struct {
		DSN            string
		MaxOpenConns   int32         `yaml:"max_open_conns"`
		MaxIdleConns   int32         `yaml:"max_idle_conns"`
		MaxIdleTime    time.Duration `yaml:"max_idle_time"`
		ConnAttempts   int32         `yaml:"conn_attempts"`
		BaseRetryDelay time.Duration `yaml:"base_retry_delay"`
		MaxRetryDelay  time.Duration `yaml:"max_retry_delay"`
	} `yaml:"postgres"`
}

type Auth struct {
	TokenSecret string        `yaml:"token_secret" env:"MEDCORE_AUTH_TOKEN_SECRET" env-default:"dev-secret-change-me"`
	TokenTTL    time.Duration `yaml:"token_ttl" env:"MEDCORE_AUTH_TOKEN_TTL" env-default:"24h"`
}

type ML struct {
	BaseURL string        `yaml:"base_url" env:"MEDCORE_ML_BASE_URL" env-default:"http://localhost:8000"`
	Timeout time.Duration `yaml:"timeout" env:"MEDCORE_ML_TIMEOUT" env-default:"3s"`
}

func MustLoad() *Config {
	loadEnvFile(".env")
	loadEnvFile("../.env")

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		panic("Load config path is failed")
	}

	//panic("Load config failed")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic(fmt.Sprintf("config file %s does not exists", configPath))

	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Sprintf("Parse config is failed: %s", err.Error()))
	}

	user := os.Getenv("MEDCORE_POSTGRES_USER")
	if user == "" {
		panic("Load DSN is failed")
	}

	pass := os.Getenv("MEDCORE_POSTGRES_PASSWORD")
	if pass == "" {
		panic("Load DSN is failed")
	}

	db := os.Getenv("MEDCORE_POSTGRES_DB")
	if db == "" {
		panic("Load DSN is failed")
	}

	port := os.Getenv("MEDCORE_POSTGRES_PORT")
	if port == "" {
		panic("Load DSN is failed")
	}

	cfg.StorageConfig.Postgres.DSN = fmt.Sprintf("host=localhost user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", user, pass, db, port)
	return &cfg
}

func loadEnvFile(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if key == "" {
			continue
		}
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}
