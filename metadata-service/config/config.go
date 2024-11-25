package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config is the configuration for the control-plane
	Config struct {
		once     sync.Once
		instance *Config
		App      `yaml:"app"`
		Etcd     `yaml:"etcd"`
		Kafka    `yaml:"kafka"`
		GRPC     `yaml:"grpc"`
		Log      `yaml:"log"`
	}

	App struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	}

	Etcd struct {
		Endpoints   []string `yaml:"endpoints" env:"ENDPOINTS"`
		DialTimeout int      `yaml:"dial_timeout" env:"DIAL_TIMEOUT"`
		TTL         int      `yaml:"ttl" env:"TTL"`
	}

	Kafka struct {
		Brokers                 []string `yaml:"brokers"`
		TopicFileMetadataSave   string   `yaml:"topic_file_metadata_save"`
		TopicFileMetadataRead   string   `yaml:"topic_file_metadata_read"`
		TopicFileMetadataDelete string   `yaml:"topic_file_metadata_delete"`
	}

	GRPC struct {
		Port string `yaml:"port"`
	}

	Log struct {
		Level string `yaml:"log_level"`
	}
)

func NewConfig() *Config {
	cfg := &Config{}

	configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatalf("config path is not set")
	}

	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		log.Fatalf("failed to get absolute path: %s, path: %s", err, configPath)
	}
	err = cleanenv.ReadConfig(absConfigPath, cfg)
	if err != nil {
		log.Fatalf("failed to get absolute path: %s, path: %s", err, configPath)
	}
	err = cleanenv.ReadConfig(configPath, cfg)
	if err != nil {
		log.Fatalf("config error: %s", err)
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("config error: %s", err)
	}

	return cfg
}

func (cfg *Config) GetInstance() *Config {
	cfg.once.Do(func() {
		cfg.instance = NewConfig()
	})

	return cfg.instance
}

func (cfg *Config) String() string {
	return fmt.Sprintf("App: %s, Version: %s, GRPC Port: %s",
		cfg.App.Name,
		cfg.App.Version,
		cfg.GRPC.Port,
	)
}

var ConfigSingleton Config