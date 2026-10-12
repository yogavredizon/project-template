package conf

import "github.com/spf13/viper"

type Config struct {
	AppName  string `mapstructure:"APP_NAME"`
	RabbitMQ struct {
		URL             string `mapstructure:"RABBITMQ_URL"`
		MaxAttempts     int    `mapstructure:"RABBITMQ_MAX_ATTEMPS"`
		ReconnectDelay  int    `mapstructure:"RABBITMQ_RECONNECT_DELAY"`
		ChannelPoolSize int    `mapstructure:"RABBITMQ_CHANNEL_POOL_SIZE"`
		PublishTimeout  int    `mapstructure:"RABBITMQ_PUBLISH_TIMEOUT"`
	}
}

func New() *Config {
	viper.SetConfigFile(".env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		panic(err)
	}

	return &config
}
