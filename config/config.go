// config/config.go
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	DBPathEmails     string
	DBPathAuth       string
	DBPathMgmt       string
	EmlStorageDir    string
	AttachStorageDir string
	LogStorageDir    string
	PortWebPortal    string
	PortAgentAPI     string
	PortSysMgr       string
}

func LoadConfig() *AppConfig {
	if err := godotenv.Load(); err != nil {
		log.Println("[CONFIG] No .env file found, relying on system environment variables.")
	}

	return &AppConfig{
		DBPathEmails:     getEnv("DB_EMAILS", "data/emails.sqlite"),
		DBPathAuth:       getEnv("DB_AUTH", "data/auth.sqlite"),
		DBPathMgmt:       getEnv("DB_MGMT", "data/mgmt.sqlite"),
		EmlStorageDir:    getEnv("DIR_EMLS", "data/storage/emls"),
		AttachStorageDir: getEnv("DIR_ATTACH", "data/storage/attachments"),
		LogStorageDir:    getEnv("DIR_LOGS", "data/storage/logs"),
		PortWebPortal:    getEnv("PORT_WEB", ":8080"),
		PortAgentAPI:     getEnv("PORT_API", ":8081"),
		PortSysMgr:       getEnv("PORT_SYSMGR", ":8088"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}