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
	Notifier      Notifier      `yaml:"notifier"`
	FHIR          FHIR          `yaml:"fhir"`
	Jobs          Jobs          `yaml:"jobs"`
	Bootstrap     Bootstrap     `yaml:"bootstrap"`
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
	BaseURL             string        `yaml:"base_url" env:"MEDCORE_ML_BASE_URL" env-default:"http://localhost:8000"`
	Timeout             time.Duration `yaml:"timeout" env:"MEDCORE_ML_TIMEOUT" env-default:"3s"`
	DatasetServiceToken string        `yaml:"dataset_service_token" env:"MEDCORE_ML_DATASET_SERVICE_TOKEN"`
	AdminToken          string        `yaml:"admin_token" env:"MEDCORE_ML_ADMIN_TOKEN"`
}

type Notifier struct {
	BaseURL        string        `yaml:"base_url" env:"MEDCORE_NOTIFIER_BASE_URL" env-default:"http://localhost:8082"`
	Timeout        time.Duration `yaml:"timeout" env:"MEDCORE_NOTIFIER_TIMEOUT" env-default:"3s"`
	Enabled        bool          `yaml:"enabled" env:"MEDCORE_NOTIFIER_ENABLED" env-default:"true"`
	CallbackSecret string        `yaml:"callback_secret" env:"MEDCORE_NOTIFIER_CALLBACK_SECRET" env-default:"dev-notifier-callback-secret"`
}

type FHIR struct {
	Enabled     bool          `yaml:"enabled" env:"MEDCORE_FHIR_ENABLED" env-default:"false"`
	Code        string        `yaml:"code" env:"MEDCORE_FHIR_CODE" env-default:"clinic-fhir"`
	Name        string        `yaml:"name" env:"MEDCORE_FHIR_NAME" env-default:"Clinic FHIR R4"`
	BaseURL     string        `yaml:"base_url" env:"MEDCORE_FHIR_BASE_URL"`
	Timeout     time.Duration `yaml:"timeout" env:"MEDCORE_FHIR_TIMEOUT" env-default:"5s"`
	BearerToken string        `yaml:"bearer_token" env:"MEDCORE_FHIR_BEARER_TOKEN"`
}

type Jobs struct {
	AutoNoShow AutoNoShowJob `yaml:"auto_no_show"`
	MLTraining MLTrainingJob `yaml:"ml_training"`
}

type Bootstrap struct {
	Admin BootstrapAdmin `yaml:"admin"`
}

type BootstrapAdmin struct {
	Enabled  bool   `yaml:"enabled" env:"MEDCORE_BOOTSTRAP_ADMIN_ENABLED" env-default:"false"`
	Email    string `yaml:"email" env:"MEDCORE_BOOTSTRAP_ADMIN_EMAIL" env-default:"admin@example.com"`
	Phone    string `yaml:"phone" env:"MEDCORE_BOOTSTRAP_ADMIN_PHONE" env-default:"+79990000000"`
	Password string `yaml:"password" env:"MEDCORE_BOOTSTRAP_ADMIN_PASSWORD" env-default:"admin12345"`
}

type AutoNoShowJob struct {
	Enabled  bool          `yaml:"enabled" env:"MEDCORE_AUTO_NO_SHOW_ENABLED" env-default:"true"`
	Interval time.Duration `yaml:"interval" env:"MEDCORE_AUTO_NO_SHOW_INTERVAL" env-default:"1m"`
	Grace    time.Duration `yaml:"grace" env:"MEDCORE_AUTO_NO_SHOW_GRACE" env-default:"30m"`
	Limit    int           `yaml:"limit" env:"MEDCORE_AUTO_NO_SHOW_LIMIT" env-default:"100"`
}

type MLTrainingJob struct {
	Enabled       bool          `yaml:"enabled" env:"MEDCORE_ML_TRAINING_ENABLED" env-default:"false"`
	Interval      time.Duration `yaml:"interval" env:"MEDCORE_ML_TRAINING_INTERVAL" env-default:"10s"`
	Timeout       time.Duration `yaml:"timeout" env:"MEDCORE_ML_TRAINING_TIMEOUT" env-default:"30m"`
	PythonPath    string        `yaml:"python_path" env:"MEDCORE_ML_PYTHON_PATH" env-default:"../ml-service/.venv/bin/python"`
	WorkDir       string        `yaml:"work_dir" env:"MEDCORE_ML_WORK_DIR" env-default:"../ml-service"`
	CoreURL       string        `yaml:"core_url" env:"MEDCORE_ML_CORE_URL" env-default:"http://localhost:8081"`
	DataPath      string        `yaml:"data_path" env:"MEDCORE_ML_DATA_PATH" env-default:"../healthcare_noshows.csv"`
	ModelsDir     string        `yaml:"models_dir" env:"MEDCORE_ML_MODELS_DIR" env-default:"artifacts/models"`
	MetricsPath   string        `yaml:"metrics_path" env:"MEDCORE_ML_METRICS_PATH" env-default:"artifacts/metrics.json"`
	MaxOutputSize int           `yaml:"max_output_size" env:"MEDCORE_ML_TRAINING_MAX_OUTPUT_SIZE" env-default:"1048576"`
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
