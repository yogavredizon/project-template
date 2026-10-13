package config

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	AppName string `mapstructure:"APP_NAME"`
	AppPort string `mapstructure:"APP_PORT"`

	OtelHost string `mapstructure:"OTEL_HOST"`

	RabbitMQ struct {
		URL             string `mapstructure:"RABBITMQ_URL"`
		MaxAttempts     int    `mapstructure:"RABBITMQ_MAX_ATTEMPTS"`
		ReconnectDelay  int    `mapstructure:"RABBITMQ_RECONNECT_DELAY"`
		ChannelPoolSize int    `mapstructure:"RABBITMQ_CHANNEL_POOL_SIZE"`
		PublishTimeout  int    `mapstructure:"RABBITMQ_PUBLISH_TIMEOUT"`
	}
}

func New() (*Config, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	if err := loadEnv(&config, v); err != nil {
		return nil, err
	}

	return &config, nil
}

func loadEnv(dst any, v *viper.Viper) error {
	value := reflect.ValueOf(dst)

	if value.Kind() != reflect.Pointer || value.IsNil() {
		return fmt.Errorf("config must be a non-nil pointer")
	}

	return setEnv(value.Elem(), v)
}

func setEnv(value reflect.Value, v *viper.Viper) error {
	typ := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		if field.Kind() == reflect.Struct {
			if err := setEnv(field, v); err != nil {
				return err
			}
			continue
		}

		key := fieldType.Tag.Get("mapstructure")
		if key == "" || key == "-" {
			continue
		}

		if !v.IsSet(key) {
			continue
		}

		raw := v.Get(key)

		switch field.Kind() {
		case reflect.String:
			field.SetString(fmt.Sprint(raw))
		case reflect.Int, reflect.Int8, reflect.Int16,
			reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(fmt.Sprint(raw), 10, field.Type().Bits())
			if err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
			field.SetInt(n)
		default:
			return fmt.Errorf("unsupported config type %s for %s", field.Type(), key)
		}
	}

	return nil
}
