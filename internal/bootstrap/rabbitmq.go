package bootstrap

import (
	"github.com/yogavredizon/log-analyzer/internal/config"
	"github.com/yogavredizon/log-analyzer/internal/sdk/rabbitmq"
)

func NewRabbitMQ(cfg *config.Config) (*rabbitmq.Connection, *rabbitmq.Publisher, error) {
	conn := rabbitmq.NewConnection(cfg)
	publisher, err := rabbitmq.NewPublisher(conn)
	if err != nil {
		conn.Shutdown()
		return nil, nil, err
	}

	return conn, publisher, nil
}
