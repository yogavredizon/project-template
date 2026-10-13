package service

import (
	"context"

	"github.com/yogavredizon/log-analyzer/internal/module/activity/dto"
	"github.com/yogavredizon/log-analyzer/internal/sdk/rabbitmq"
)

type Activity struct {
	publisher *rabbitmq.Publisher
}

func NewActivity(publisher *rabbitmq.Publisher) *Activity {
	return &Activity{
		publisher: publisher,
	}
}

func (a *Activity) Publish(ctx context.Context, req dto.Request) error {
	if err := a.publisher.Publish(ctx, rabbitmq.Message{Exchange: "events.activity", RoutingKey: "events.activity.key", Headers: nil, Priority: 1, Body: req}); err != nil {
		return err
	}

	return nil
}
