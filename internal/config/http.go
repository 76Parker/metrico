// Пакет содержит в себе конфигурацию HTTP-сервера
package config

import "time"

type HTTP struct {
	Address           string        `yaml:"address"`
	HashKey           string        `yaml:"hash_key"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	MaxHeaderBytes    int           `yaml:"max_header_bytes"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
}

type Postgres struct {
	DSN string
}

type SnapshotService struct {
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
}
