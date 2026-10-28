package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN,required"`
	TelegramChatID   string `env:"TELEGRAM_CHAT_ID,required"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("Failed to parse environment configuration: %w", err)
	}
	return cfg, nil
}
