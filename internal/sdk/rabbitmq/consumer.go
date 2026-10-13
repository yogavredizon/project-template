package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler func(ctx context.Context, msg Delivery) error

type ConsumerConfig struct {
	QueueName   string
	ConsumerTag string
	Exchange    string
	RoutingKey  string

	Prefetch int
	Worker   int

	Retriable     bool
	MaxRetry      int
	DLQExchange   string
	DLQRoutingKey string

	// LastDLQ will store failed consumed message that reach the maximum retry
	LastDLQExchange   string
	LastDLQRoutingKey string
}

type Consumer struct {
	conn      *Connection
	publisher *Publisher
	cfg       *ConsumerConfig
	wg        sync.WaitGroup
	handler   Handler
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewConsumer(conn *Connection, cfg *ConsumerConfig, handler Handler, publisher *Publisher) (*Consumer, error) {
	if conn == nil {
		return nil, fmt.Errorf("RabbitMQ is not yet ready to accept connection")
	}

	if cfg.Worker <= 0 {
		cfg.Worker = 1
	}

	ctx, cancel := context.WithCancel(context.Background())

	consumer := &Consumer{
		conn:      conn,
		cfg:       cfg,
		handler:   handler,
		publisher: publisher,
		ctx:       ctx,
		cancel:    cancel,
	}

	if err := consumer.declareDLQ(context.Background(), "topic"); err == nil {
		log.Printf("failed to declare dead letter queue. error: %v", err)
	}

	return consumer, nil
}

func (c *Consumer) Consume(ctx context.Context) error {
	if err := c.conn.WaitReady(ctx); err != nil {
		return err
	}

	notify := c.conn.subscribe()

	for {
		if err := c.subscribe(); err != nil {
			select {
			case <-ctx.Done():
				return nil
			case <-c.ctx.Done():
				return nil
			case <-time.After(time.Duration(c.conn.cfg.RabbitMQ.ReconnectDelay) * time.Second):
				continue
			}
		}

		select {
		case <-notify:
		case <-ctx.Done():
			c.cancel()
			c.wg.Wait()
			return nil
		case <-c.ctx.Done():
			c.wg.Wait()
			return nil
		}
	}
}

func (c *Consumer) subscribe() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel, error : %w", err)
	}

	if err := ch.Qos(c.cfg.Prefetch, 0, false); err != nil {
		_ = ch.Close()
		return err
	}

	delivery, err := ch.Consume(c.cfg.QueueName, c.cfg.ConsumerTag, false, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		return err
	}

	for range c.cfg.Worker {
		c.wg.Add(1)
		go c.worker(delivery)
	}

	return nil
}

func (c *Consumer) worker(ch <-chan amqp.Delivery) {
	defer c.wg.Done()

	for delivery := range ch {
		c.dispatch(delivery)
	}
}

func (c *Consumer) dispatch(delivery amqp.Delivery) {
	if err := c.handler(context.Background(), Delivery{Body: delivery.Body, ContentType: delivery.ContentType}); err != nil {
		if !c.cfg.Retriable {
			if err := delivery.Reject(false); err != nil {
				log.Printf("failed to acknowledge message. error: %v", err)
			}
			return
		}

		if err := c.retryBackoff(context.Background(), delivery); err != nil {
			log.Printf("failed to retry process. error: %v", err)
		}
		return
	}

	delivery.Ack(false)
}

func (c *Consumer) retryBackoff(ctx context.Context, delivery amqp.Delivery) error {
	if c.cfg.DLQExchange == "" {
		return delivery.Reject(false)
	}

	headers := delivery.Headers
	if headers == nil {
		headers = make(map[string]any)
	}

	var retryCount int = 0
	if val, ok := headers["x-retry-count"]; ok {
		switch v := val.(type) {
		case int:
			retryCount = v
		case int32:
			retryCount = int(v)
		case int64:
			retryCount = int(v)
		}
	}

	retryCount++

	headers["x-retry-count"] = retryCount

	delay := backoff(retryCount, time.Second*5)
	exchange := c.cfg.DLQExchange
	routingKey := fmt.Sprintf("%s.%ds", c.cfg.DLQRoutingKey, int(delay.Seconds()))

	if c.cfg.MaxRetry > 0 && retryCount > c.cfg.MaxRetry {
		if c.cfg.LastDLQExchange == "" {
			return delivery.Reject(false)
		}

		headers["x-error-message"] = "max retry has reached"
		exchange = c.cfg.LastDLQExchange
		routingKey = c.cfg.LastDLQRoutingKey
	}

	if err := c.publisher.Publish(ctx,
		Message{
			Exchange:   exchange,
			RoutingKey: routingKey,
			Headers:    headers,
			Body:       Delivery{Body: delivery.Body, ContentType: delivery.ContentType},
		},
	); err != nil {
		return delivery.Nack(false, true)
	}

	return delivery.Ack(false)
}

func (c *Consumer) declareDLQ(ctx context.Context, kind string) error {
	if !c.cfg.Retriable && c.cfg.DLQExchange == "" {
		return nil
	}

	ch, err := c.publisher.acquireChannel(ctx)
	if err != nil {
		return err
	}

	defer c.publisher.releaseChannel(ch)

	maxRetry := c.cfg.MaxRetry
	if maxRetry > 3 {
		maxRetry = 3
	}

	err = ch.ExchangeDeclare(c.cfg.DLQExchange, kind, true, false, false, false, nil)
	if err != nil {
		return err
	}

	for i := range maxRetry {
		delay := backoff(i+1, time.Second*5)

		key := fmt.Sprintf("%s.%ds", c.cfg.DLQRoutingKey, int(delay.Seconds()))
		dlq := fmt.Sprintf("%s.%s", c.cfg.QueueName, delay.String())

		queue, err := ch.QueueDeclare(dlq, true, false, false, false,
			amqp.Table{
				"x-dead-letter-exchange":    c.cfg.Exchange,
				"x-dead-letter-routing-key": c.cfg.RoutingKey,
				"x-message-ttl":             int32(delay.Milliseconds()),
			},
		)

		if err != nil {
			continue
		}

		if err := ch.QueueBind(queue.Name, key, c.cfg.DLQExchange, false, nil); err != nil {
			continue
		}
	}

	return nil
}
