package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GeminiAPIKey        string
	DocxTemplatePath    string
	DocxNegTemplatePath string
	ServerAddr          string
	RedisAddr           string
	DatabaseURL         string
}

func Load() Config {
	_ = godotenv.Load()

	templatePath := os.Getenv("DOCX_TEMPLATE_PATH")
	if templatePath == "" {
		templatePath = "2INT.docx"
	}

	negTemplatePath := os.Getenv("DOCX_NEG_TEMPLATE_PATH")
	if negTemplatePath == "" {
		negTemplatePath = "3NEG.docx"
	}

	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	return Config{
		GeminiAPIKey:        os.Getenv("GEMINI_API_KEY"),
		DocxTemplatePath:    templatePath,
		DocxNegTemplatePath: negTemplatePath,
		ServerAddr:          addr,
		RedisAddr:           redisAddr,
		DatabaseURL:         os.Getenv("DATABASE_URL"),
	}
}
