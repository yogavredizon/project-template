package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yogavredizon/log-analyzer/internal/conf"
)

const (
	maxReconnectAttempts = 5
	maxReconnectDelay    = 3
)

type Connection struct {
	cfg  conf.Config
	mu   sync.RWMutex
	conn *amqp.Connection

	subs   []chan struct{}
	subsMu sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	ready  chan struct{}
	once   sync.Once
}

func NewConnection(cfg conf.Config) *Connection {
	ctx, cancel := context.WithCancel(context.Background())

	if cfg.RabbitMQ.MaxAttempts <= 0 {
		cfg.RabbitMQ.MaxAttempts = maxReconnectAttempts
	}

	if cfg.RabbitMQ.ReconnectDelay <= 0 {
		cfg.RabbitMQ.ReconnectDelay = maxReconnectDelay
	}

	c := &Connection{
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
		ready:  make(chan struct{}),
	}

	go c.reconnectLoop()

	return c
}

func (c *Connection) Channel() (*amqp.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil || c.conn.IsClosed() {
		return nil, fmt.Errorf("rabbitmq connection is not ready")
	}

	return c.conn.Channel()
}

func (c *Connection) Ready() <-chan struct{} {
	return c.ready
}

func (c *Connection) WaitReady(ctx context.Context) error {
	select {
	case <-c.ready:
		return nil
	case <-c.ctx.Done():
		return fmt.Errorf("rabbitmq connection failed: max reconnect attempts reached")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Connection) IsHealthy() bool {
	if c == nil {
		return false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.conn != nil && !c.conn.IsClosed()
}

func (c *Connection) subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	c.subsMu.Lock()
	c.subs = append(c.subs, ch)
	c.subsMu.Unlock()

	return ch
}

func (c *Connection) notifySubsriber() {
	c.subsMu.RLock()
	defer c.subsMu.RUnlock()

	for _, sub := range c.subs {
		select {
		case sub <- struct{}{}:
		default:
		}
	}
}

func (c *Connection) reconnectLoop() {
	attempts := 0
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		if c.cfg.RabbitMQ.MaxAttempts > 0 && attempts >= c.cfg.RabbitMQ.MaxAttempts {
			log.Printf("RabbitMQ max reconect attempt is reached. Attempt : %d", attempts)
			c.cancel()
			return
		}

		conn, err := amqp.Dial(c.cfg.RabbitMQ.URL)
		if err != nil {
			attempts++

			delay := backoff(attempts, time.Duration(c.cfg.RabbitMQ.ReconnectDelay)*time.Second)
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(delay):
				continue
			}
		}

		attempts = 0

		c.mu.Lock()
		c.conn = conn
		c.mu.Unlock()

		c.once.Do(func() { close(c.ready) })

		log.Println("RabbitMQ is ready!")

		c.notifySubsriber()

		closeCh := conn.NotifyClose(make(chan *amqp.Error, 1))
		select {
		case <-c.ctx.Done():
			return
		case err := <-closeCh:
			if err != nil {
				log.Printf("RabbitMQ connection is closed. error : %v, reconnecting...", err)
			}
		}

	}
}

func backoff(attempt int, base time.Duration) time.Duration {
	if attempt <= 0 {
		return base
	}

	if attempt > 30 {
		return 30 * time.Second
	}

	d := base * time.Duration(1<<uint8(attempt-1))
	if d > 30*time.Second {
		d = 30 * time.Second
	}

	return d
}

func (c *Connection) Shutdown() {
	c.cancel()

	c.subsMu.Lock()
	for _, sub := range c.subs {
		close(sub)
	}
	c.subs = nil
	c.subsMu.Unlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil && !c.conn.IsClosed() {
		if err := c.conn.Close(); err != nil {
			log.Printf("failed to close RabbitMQ connection. Error: %v", err)
		}
	}
}
