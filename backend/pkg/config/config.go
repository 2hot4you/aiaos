package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server     ServerConfig
	DB         DBConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Encryption EncryptionConfig
}

type ServerConfig struct {
	Port string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret     string
	ExpireDays int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type EncryptionConfig struct {
	Key string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "aiaos"),
			Password: getEnv("DB_PASSWORD", "aiaos_password"),
			DBName:   getEnv("DB_NAME", "aiaos"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "aiaos-dev-jwt-secret-key-at-least-32chars"),
			ExpireDays: getEnvInt("JWT_EXPIRE_DAYS", 7),
		},
		Encryption: EncryptionConfig{
			Key: getEnv("ENCRYPTION_KEY", "aiaos-dev-encryption-key-32bytes!"),
		},
	}
}

func (c *DBConfig) DSN() string {
	return "host=" + c.Host +
		" port=" + c.Port +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.DBName +
		" sslmode=" + c.SSLMode +
		" TimeZone=UTC"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
