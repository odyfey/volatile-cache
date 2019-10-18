package config

import (
	"os"
	"strconv"
	"sync"
)

type Config struct {
	Host           string
	Port           int
	ExpirationTime int // in seconds
	BucketsNum     int
	StoreFileName  string
}

var instance *Config
var once sync.Once

func GetInstance() *Config {
	once.Do(func() {
		instance = &Config{
			Host:           getEnv("HOST", "localhost"),
			Port:           getEnvAsInt("PORT", 20153),
			ExpirationTime: getEnvAsInt("EXP_TIME", 60),
			BucketsNum:     getEnvAsInt("BKT_NUM", 256),
			StoreFileName:  getEnv("STORE_FILENAME", "store.dat"),
		}
	})
	return instance
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(name string, defaultValue int) int {
	envValue := getEnv(name, "")
	if value, err := strconv.Atoi(envValue); err == nil {
		return value
	}
	return defaultValue
}
