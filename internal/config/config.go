package config

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/MaiM-with-u/maibot-telegram-adapter/internal/logger"
	"sync"
)

var (
	cfg  *Config
	once sync.Once
)

type MaibotConfig struct {
	URL string
}

type MessageFilter struct {
	Mode string
	List []int64
}

type MessageFilterConfig struct {
	BannedUsers []int64
	Groups      MessageFilter
	Private     MessageFilter
}

type Config struct {
	Platform         string
	TelegramBotToken string
	MaiBot           MaibotConfig
	MessageFilter    MessageFilterConfig
}

func NewDefaultConfig() *Config {
	return &Config{
		Platform:         "telegram",
		TelegramBotToken: "",
		MaiBot: MaibotConfig{
			URL: "ws://localhost:8080",
		},
		MessageFilter: MessageFilterConfig{
			BannedUsers: []int64{},
			Groups: MessageFilter{
				Mode: "whitelist",
				List: []int64{},
			},
			Private: MessageFilter{
				Mode: "blacklist",
				List: []int64{},
			},
		},
	}
}

func Load() *Config {
	configPath := *flag.String("config", "config.toml", "path to config file")
	if configPath == "" {
		configPath = "config.toml"
	}

	once.Do(func() {
		c := NewDefaultConfig() // 初始化默认值
		if _, err := toml.DecodeFile(configPath, c); err != nil {
			logger.Fatalf("failed to load config: %v", err)
		}
		cfg = c
	})
	return cfg
}

func Get() *Config {
	if cfg == nil {
		logger.Fatal("config not initialized: call LoadConfig first")
	}
	return cfg
}
