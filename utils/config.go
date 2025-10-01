package utils

import (
	"bufio"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Port        string
	ChunkSize   int
	MaxRoutine  uint
	Key         byte
	WebhookPath string
	UserToken   string
}

var cfg *Config = nil
var webhooks []string

func createDefaultCfg() {
	cfg := Config{
		Port:        "8080",
		ChunkSize:   1023 * 1024 * 10,
		MaxRoutine:  6,
		Key:         byte(0x69),
		WebhookPath: "/path/to/file.txt",
		UserToken:   "<Discord user token>",
	}

	b, err := toml.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	_ = os.WriteFile("config.toml", b, 0644)
}

func ReadCfg() Config {
	if _, err := os.Stat("config.toml"); err != nil {
		createDefaultCfg()
	}

	if cfg == nil {
		var temp Config
		data, err := os.ReadFile("config.toml")
		if err != nil {
			panic(err)
		}

		err = toml.Unmarshal(data, &temp)
		if err != nil {
			panic(err)
		}
		cfg = &temp
	}
	return *cfg
}

func (cfg Config) LoadWebhooks() {
	file, err := os.Open(cfg.WebhookPath)
	Check(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		webhooks = append(webhooks, scanner.Text())
	}
}

func (cfg Config) GetWebhooks() []string {
	if len(webhooks) == 0 {
		cfg.LoadWebhooks()
	}
	return webhooks
}
