// Пакет содержит в себе конфигурацию HTTP-сервера
package config

import "time"

type HTTP struct {
	Address           string        `yaml:"address"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	MaxHeaderBytes    int           `yaml:"max_header_bytes"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
}
