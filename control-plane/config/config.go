package config

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config is the configuration for the control-plane
	Config struct {
		once        sync.Once
		instance    *Config
		App         `yaml:"app"`
		GRPC        `yaml:"grpc"`
		FileStorage `yaml:"file_storage"`
		Log         `yaml:"log"`
	}

	App struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	}

	GRPC struct {
		Port string `yaml:"port"`
	}

	FileStorage struct {
		Chunk_size uint32 `yaml:"chunk_size"`
	}

	Log struct {
		Level string `yaml:"log_level"`
	}
)

func NewConfig() *Config {
	cfg := &Config{}

	configPath, err := filepath.Abs("./config/config.yml")
	if err != nil {
		log.Fatalf("failed to get absolute path: %s", err)
	}
	err = cleanenv.ReadConfig(configPath, cfg)
	if err != nil {
		log.Fatalf("config error: %s", err)
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("config error: %s", err)
	}

	cfg.FileStorage.Chunk_size = cfg.FileStorage.Chunk_size * 1024 * 1024

	return cfg
}

func (cfg *Config) GetInstance() *Config {
	cfg.once.Do(func() {
		cfg.instance = NewConfig()
	})

	return cfg.instance
}

func (cfg *Config) String() string {
	return fmt.Sprintf("App: %s, Version: %s, GRPC Port: %s, Chunk Size (MB): %d",
		cfg.App.Name,
		cfg.App.Version,
		cfg.GRPC.Port,
		cfg.FileStorage.Chunk_size)
}

var ConfigSingleton Config
