// Package config는 설정 파일을 읽고 파싱하는 기능을 제공합니다.
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"space/internal/domain"
)

// Config는 애플리케이션 전체 설정을 담는 구조체입니다.
type Config struct {
	Server    ServerConfig     `toml:"server"`
	Databases []DatabaseConfig `toml:"databases"`
	Logging   LoggingConfig    `toml:"logging"`
}

// ServerConfig는 서버 설정입니다.
type ServerConfig struct {
	Port            string   `toml:"port"`
	ShutdownTimeout string   `toml:"shutdown_timeout"`
	AllowedOrigins  []string `toml:"allowed_origins"`
}

// DatabaseConfig는 개별 데이터베이스 설정입니다.
type DatabaseConfig struct {
	ID                string `toml:"id"`
	Name              string `toml:"name"`
	Type              string `toml:"type"` // "postgresql", "oracle19c", "mariadb"
	Host              string `toml:"host"`
	Port              int    `toml:"port"`
	Username          string `toml:"username"`
	Password          string `toml:"password"`
	Schema            string `toml:"schema"`
	ConnectOnStartup  bool   `toml:"connect_on_startup"`
	ConnectionTimeout string `toml:"connection_timeout"` // "60s"
}

// LoggingConfig는 로깅 설정입니다.
type LoggingConfig struct {
	Level  string `toml:"level"`  // debug, info, warn, error
	Prefix string `toml:"prefix"` // "[DMS]"
}

// Load는 지정된 경로의 TOML 파일을 읽어 Config 구조체를 반환합니다.
func Load(configPath string) (*Config, error) {
	// 파일 존재 확인
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	var config Config

	// TOML 파일 파싱
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 기본값 설정
	if config.Server.Port == "" {
		config.Server.Port = "8080"
	}
	if config.Server.ShutdownTimeout == "" {
		config.Server.ShutdownTimeout = "5s"
	}
	if config.Logging.Prefix == "" {
		config.Logging.Prefix = "[DMS]"
	}
	// AllowedOrigins 기본값 ← 추가!
	if len(config.Server.AllowedOrigins) == 0 {
		config.Server.AllowedOrigins = []string{"*"} // 개발용 기본값
	}

	return &config, nil
}

// GetShutdownTimeout은 shutdown_timeout을 time.Duration으로 변환합니다.
func (c *ServerConfig) GetShutdownTimeout() time.Duration {
	duration, err := time.ParseDuration(c.ShutdownTimeout)
	if err != nil {
		return 5 * time.Second // 기본값
	}
	return duration
}

// GetConnectionTimeout은 connection_timeout을 time.Duration으로 변환합니다.
func (d *DatabaseConfig) GetConnectionTimeout() time.Duration {
	duration, err := time.ParseDuration(d.ConnectionTimeout)
	if err != nil {
		return 60 * time.Second // 기본값
	}
	return duration
}

// ToDomain은 DatabaseConfig를 domain.Database로 변환합니다.
func (d *DatabaseConfig) ToDomain() (*domain.Database, error) {
	// type 문자열을 domain.DatabaseType으로 변환
	var dbType domain.DatabaseType
	switch d.Type {
	case "postgresql":
		dbType = domain.PostgreSQL
	case "oracle19c":
		dbType = domain.Oracle19c
	case "mariadb":
		dbType = domain.MariaDB
	default:
		return nil, fmt.Errorf("unsupported database type: %s", d.Type)
	}

	return &domain.Database{
		ID:       d.ID,
		Name:     d.Name,
		Type:     dbType,
		Host:     d.Host,
		Port:     d.Port,
		Username: d.Username,
		Password: d.Password,
		Schema:   d.Schema,
		Status:   domain.Disconnected,
	}, nil
}
