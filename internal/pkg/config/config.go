package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string     `yaml:"env" env:"ENV" env-required:"true"`
	DB         DB         `yaml:"db"`
	Server     Server     `yaml:"server"`
	GRPCServer GRPCServer `yaml:"grpcServer"`
	Kafka      Kafka      `yaml:"kafka"`
	Bot        Bot        `yaml:"bot"`
}

type DB struct {
	Username string `yaml:"username" env-required:"true"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD"`
	Host     string `yaml:"host" env-required:"true"`
	Port     string `yaml:"port" env-required:"true"`
	DB       string `yaml:"dbType"`
	Reload   bool   `yaml:"reload"`
	Version  int64  `yaml:"version"`
}

type Server struct {
	Host            string `yaml:"host"`
	Port            string `yaml:"port"`
	ShutDownTimeout int64  `yaml:"shutdown"`
}

type GRPCServer struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Kafka struct {
	Brokers           []string `yaml:"brokers"`
	Topic             string   `yaml:"topic"`
	Partitions        int      `yaml:"partitions"`
	ReplicationFactor int      `yaml:"replication"`
	Group             string   `yaml:"group"`
}

type Bot struct {
	Host  string `yaml:"host"`
	Token string `env:"TOKEN" env-required:"true"`
}

func New(configPath string) (Config, error) {
	var cfg Config
	if err := godotenv.Load(); err != nil {
		return cfg, err
	}
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
