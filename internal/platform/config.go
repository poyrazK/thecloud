// Package platform provides platform-specific configurations and utilities.
package platform

import (
	"os"

	"github.com/joho/godotenv"
)

// Config stores application configuration loaded from environment variables.
type Config struct {
	Port                 string
	DatabaseURL          string
	DatabaseReadURL      string
	Environment          string
	SecretsEncryptionKey string
	ComputeBackend       string
	NetworkBackend       string
	DefaultVPCCIDR       string
	NetworkPoolStart     string
	NetworkPoolEnd       string
	DBMaxConns           string
	DBMinConns           string
	RedisURL             string
	RateLimitGlobal      string
	RateLimitAuth        string
	StorageBackend       string
	// StorageSecret is the secret key used for signing presigned URLs
	StorageSecret      string
	LvmVgName          string
	ObjectStorageMode  string
	ObjectStorageNodes string
}

// NewConfig loads configuration from the environment with defaults.
func NewConfig() (*Config, error) {
	_ = godotenv.Load() // Ignore error if .env doesn't exist

	return &Config{
		Port:                 getEnv("PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://cloud:cloud@localhost:5433/thecloud"),
		DatabaseReadURL:      getEnv("DATABASE_READ_URL", ""), // Default to empty (use primary)
		Environment:          getEnv("APP_ENV", "development"),
		SecretsEncryptionKey: os.Getenv("SECRETS_ENCRYPTION_KEY"),
		ComputeBackend:       getEnv("COMPUTE_BACKEND", "docker"),
		NetworkBackend:       getEnv("NETWORK_BACKEND", "ovs"),
		DefaultVPCCIDR:       getEnv("DEFAULT_VPC_CIDR", "10.0.0.0/16"),
		NetworkPoolStart:     getEnv("NETWORK_POOL_START", "192.168.100.0"),
		NetworkPoolEnd:       getEnv("NETWORK_POOL_END", "192.168.200.255"),
		DBMaxConns:           getEnv("DB_MAX_CONNS", "20"),
		DBMinConns:           getEnv("DB_MIN_CONNS", "2"),
		RedisURL:             getEnv("REDIS_URL", "localhost:6379"),
		RateLimitGlobal:      getEnv("RATE_LIMIT_GLOBAL", "100"),
		RateLimitAuth:        getEnv("RATE_LIMIT_AUTH", "10"),
		StorageBackend:       getEnv("STORAGE_BACKEND", "noop"),
		StorageSecret:        getEnv("STORAGE_SECRET", "storage-secret-key"),
		LvmVgName:            getEnv("LVM_VG_NAME", "thecloud-vg"),
		ObjectStorageMode:    getEnv("OBJECT_STORAGE_MODE", "local"),
		ObjectStorageNodes:   getEnv("OBJECT_STORAGE_NODES", ""),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
