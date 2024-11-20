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
		once        sync.Once
		instance    *Config
		App         `yaml:"app"`
		Consensus   `yaml:"consensus"`
		GRPC        `yaml:"grpc"`
		FileStorage `yaml:"file_storage"`
		Log         `yaml:"log"`
	}

	App struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	}

	Consensus struct {
		NodeId    string   `yaml:"node_id" env:"NODE_ID"`
		Followers []string `yaml:"followers" env:"FOLLOWERS"`
		RaftPort  uint16   `yaml:"raft_port"`
		RaftDir   string   `yaml:"raft_dir"`
	}

	GRPC struct {
		Port string `yaml:"port"`
	}

	FileStorage struct {
		Chunk_size uint32 `yaml:"chunk_size"`
		Saga_ttl   uint32 `yaml:"saga_ttl"`
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
	return fmt.Sprintf("App: %s, Version: %s, GRPC Port: %s, Chunk Size (MB): %d, Node ID: %s, followers: %v",
		cfg.App.Name,
		cfg.App.Version,
		cfg.GRPC.Port,
		cfg.FileStorage.Chunk_size,
		cfg.Consensus.NodeId,
		cfg.Consensus.Followers,
	)
}

var ConfigSingleton Config
