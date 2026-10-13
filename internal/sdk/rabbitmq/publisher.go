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

type Publisher struct {
	conn *Connection
	cfg  *conf.Config
	pool chan *amqp.Channel
	mu   sync.Mutex
}

func NewPublisher(conn *Connection) (*Publisher, error) {
	if conn == nil {
		return nil, fmt.Errorf("RabbitMQ is not yet ready to publish a message")
	}

	cfg := conn.cfg

	poolSize := cfg.RabbitMQ.ChannelPoolSize
	if poolSize <= 0 {
		poolSize = 20
	}

	cfg.RabbitMQ.ChannelPoolSize = poolSize

	p := &Publisher{
		conn: conn,
		cfg:  cfg,
		pool: make(chan *amqp.Channel, poolSize),
	}

	notify := conn.subscribe()
	go p.watchReconnect(notify)

	if p.conn.IsHealthy() {
		p.rebuildPool()
	}

	return p, nil
}

func (p *Publisher) Publish(parent context.Context, msg Message) error {
	if p == nil || p.conn == nil {
		return fmt.Errorf("rabbitmq publisher is not initialized")
	}

	body, contentType, err := msg.Body.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode body. error: %w", err)
	}

	pub := amqp.Publishing{
		Headers:      msg.Headers,
		ContentType:  contentType,
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Priority:     msg.Priority,
		Body:         body,
	}

	ch, err := p.acquireChannel(parent)
	if err != nil {
		return err
	}

	defer p.releaseChannel(ch)

	timeout := time.Second * time.Duration(p.cfg.RabbitMQ.PublishTimeout)
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	return ch.PublishWithContext(ctx, msg.Exchange, msg.RoutingKey, false, false, pub)
}

func (p *Publisher) acquireChannel(ctx context.Context) (*amqp.Channel, error) {
	select {
	case ch := <-p.pool:
		if ch != nil && !ch.IsClosed() {
			return ch, nil
		}
	default:
	}

	if !p.conn.IsHealthy() {
		if err := p.conn.WaitReady(ctx); err != nil {
			return nil, fmt.Errorf("RabbitMQ connection not ready: %w", err)
		}
	}

	return p.conn.Channel()
}

func (p *Publisher) releaseChannel(ch *amqp.Channel) {
	if ch == nil || ch.IsClosed() {
		return
	}

	select {
	case p.pool <- ch:
	default:
		_ = ch.Close()
	}
}

func (p *Publisher) rebuildPool() {
	p.mu.Lock()
	defer p.mu.Unlock()

drainLoop:
	for {
		select {
		case ch := <-p.pool:
			if ch != nil {
				_ = ch.Close()
			}
		default:
			break drainLoop
		}
	}

	for range p.cfg.RabbitMQ.ChannelPoolSize {
		ch, err := p.conn.Channel()
		if err != nil {
			break
		}

		select {
		case p.pool <- ch:
		default:
			if err := ch.Close(); err != nil {
				log.Printf("RabbitMQ failed to close channel. error: %v", err)
			}
		}
	}
}

func (p *Publisher) watchReconnect(notify chan struct{}) {
	for range notify {
		p.rebuildPool()
	}
}
