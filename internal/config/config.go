package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

const defaultConfigPath = "configs/application.yml"

type Config struct {
	App      AppConfig      `yaml:"app"`
	Postgres PostgresConfig `yaml:"postgres"`
	Kafka    KafkaConfig    `yaml:"kafka"`
}

type AppConfig struct {
	Name string     `yaml:"name"`
	Env  string     `yaml:"env"`
	HTTP HTTPConfig `yaml:"http"`
}

type HTTPConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
}

type KafkaConfig struct {
	BootstrapServers []string          `yaml:"bootstrap_servers"`
	Topics           KafkaTopicsConfig `yaml:"topics"`
}

type KafkaTopicsConfig struct {
	GuestEvents   string `yaml:"guest_events"`
	BookingEvents string `yaml:"booking_events"`
	RoomEvents    string `yaml:"room_events"`
}

func Load(path string) (Config, error) {
	if path == "" {
		path = defaultConfigPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	cfg := defaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config %q: %w", path, err)
	}

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		App: AppConfig{
			Name: "hotel-booking",
			Env:  "local",
			HTTP: HTTPConfig{
				Host: "0.0.0.0",
				Port: 8080,
			},
		},
		Postgres: PostgresConfig{
			Host:     "localhost",
			Port:     5433,
			Database: "booking",
			Username: "booking",
			Password: "booking",
			SSLMode:  "disable",
		},
		Kafka: KafkaConfig{
			BootstrapServers: []string{"localhost:9092"},
			Topics: KafkaTopicsConfig{
				GuestEvents:   "guest-events",
				BookingEvents: "booking-events",
				RoomEvents:    "room-events",
			},
		},
	}
}

func (c Config) HTTPAddress() string {
	return fmt.Sprintf("%s:%d", c.App.HTTP.Host, c.App.HTTP.Port)
}

func (c Config) PostgresDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Postgres.Username,
		c.Postgres.Password,
		c.Postgres.Host,
		c.Postgres.Port,
		c.Postgres.Database,
		c.Postgres.SSLMode,
	)
}
