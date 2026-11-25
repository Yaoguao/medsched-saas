package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env"`
	HTTPServer struct {
		Address     string        `yaml:"address"`
		Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
		IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	} `yaml:"http_server"`
	Storage struct {
		Redis struct {
			Host     string        `yaml:"host"`
			Port     int           `yaml:"port"`
			DB       int           `yaml:"db"`
			Password string        `yaml:"password"`
			CacheTTL time.Duration `yaml:"cache_ttl"`
		} `yaml:"redis"`
	} `yaml:"storage"`
	RabbitMQ struct {
		URL            string        `yaml:"url"`
		DelayQueue     string        `yaml:"delay_queue"`
		ReadyQueue     string        `yaml:"ready_queue"`
		RetryQueue     string        `yaml:"retry_queue"`
		PrefetchCount  int           `yaml:"prefetch_count"`
		ReconnectDelay time.Duration `yaml:"reconnect_delay"`
	} `yaml:"rabbit_mq"`
	Notifier struct {
		RetryInitialDelay  time.Duration `yaml:"retry_initial_delay"`
		MaxRetries         int           `yaml:"max_retries"`
		Channels           []string      `yaml:"channels"`
		CoreCallbackURL    string        `yaml:"core_callback_url" env:"NOTIFIER_CORE_CALLBACK_URL" env-default:"http://localhost:8081/api/v1"`
		CoreCallbackSecret string        `yaml:"core_callback_secret" env:"NOTIFIER_CORE_CALLBACK_SECRET" env-default:"dev-notifier-callback-secret"`
	} `yaml:"notifier"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("Load config path is failed")
	}

	//panic("Load config failed")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file %s does not exists", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Parse config is failed: %s", err.Error())
	}
	return &cfg
}
