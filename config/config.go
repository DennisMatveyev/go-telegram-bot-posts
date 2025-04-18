package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env             string        `env:"ENV" env-default:"dev"`
	LogPath         string        `env:"LOG_PATH" env-default:""`
	DBUrl           string        `env:"DB_URL" env-required:"true"`
	TGBotToken      string        `env:"TG_BOT_TOKEN" env-required:"true"`
	TGChannelID     int64         `env:"TG_CHANNEL_ID" env-required:"true"`
	OpenAIKey       string        `env:"OPENAI_KEY" env-required:"true"`
	FetchInterval   time.Duration `env:"FETCH_INTERVAL" env-default:"30m"`
	PublishInterval time.Duration `env:"PUBLISH_INTERVAL" env-default:"30m"`
	DaysClearOld    int           `env:"DAYS_CLEAR_OLD" env-default:"1"`
}

func MustLoad() Config {
	var cfg Config
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		panic("failed to read env: " + err.Error())
	}
	return cfg
}
