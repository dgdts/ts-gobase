package rabbitmq

import (
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	address string
	conn    *amqp.Connection
	channel *amqp.Channel
}

type Producer struct {
	client *Client
	config *ProducerConf
}
type Consumer struct {
	client       *Client
	config       *ConsumerConf
	handler      func([]byte) error
	done         chan struct{}
	closeErr     chan *amqp.Error
	reconnecting bool
}

func NewClient(address string) (*Client, error) {
	conn, err := amqp.Dial(address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	return &Client{
		conn:    conn,
		channel: channel,
		address: address,
	}, nil
}

func InitProducer(conf *ProducerConf) (*Producer, error) {
	client, err := NewClient(conf.Address)
	if err != nil {
		return nil, err
	}
	if conf.Exchange != "" {
		err := client.channel.ExchangeDeclare(
			conf.Exchange,
			"direct",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to declare exchange: %v", err)
		}
	}

	if conf.Confirm {
		if err := client.channel.Confirm(false); err != nil {
			return nil, fmt.Errorf("failed to put channel in confirm mode: %v", err)
		}
	}
	return &Producer{client: client, config: conf}, nil
}

func InitConsumer(conf *ConsumerConf) (*Consumer, error) {
	if conf.PrefetchCount <= 0 {
		conf.PrefetchCount = DefaultPrefetchCount
	}
	if conf.RetryCount <= 0 {
		conf.RetryCount = DefaultRetryCount
	}
	client, err := NewClient(conf.Address)
	if err != nil {
		return nil, err
	}

	consumer := &Consumer{
		config:   conf,
		client:   client,
		done:     make(chan struct{}),
		closeErr: make(chan *amqp.Error),
	}

	if err := client.channel.Qos(conf.PrefetchCount, 0, false); err != nil {
		return nil, fmt.Errorf("failed to set QoS: %v", err)
	}
	return consumer, nil
}

func (c *Consumer) Start(handler func([]byte) error) error {
	c.handler = handler

	c.client.conn.NotifyClose(c.closeErr)

	if err := c.consume(); err != nil {
		return err
	}

	go c.monitor()

	return nil
}

// Stop 停止消费者
func (c *Consumer) Stop() {
	close(c.done)
}

func (c *Consumer) consume() error {
	messages, err := c.client.channel.Consume(
		c.config.Queue,
		"",
		c.config.AutoAck,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %v", err)
	}

	go func() {
		for {
			select {
			case <-c.done:
				return
			case msg, ok := <-messages:
				if !ok {
					return
				}
				// 处理消息
				if err := c.processMessage(msg); err != nil {
					hlog.Errorf("Error processing message: %v", err)
				}
			}
		}
	}()

	return nil
}

func (c *Consumer) processMessage(msg amqp.Delivery) error {
	var err error
	for attempt := 0; attempt <= c.config.RetryCount; attempt++ {
		err = c.handler(msg.Body)
		if err == nil {
			if !c.config.AutoAck {
				return msg.Ack(false)
			}
			return nil
		}

		hlog.Warnf("Processing attempt %d failed: %v", attempt+1, err)

		if attempt < c.config.RetryCount {
			time.Sleep(time.Duration(c.config.RetryDelay) * time.Second)
		}
	}

	if !c.config.AutoAck {
		_ = msg.Nack(false, true)
	}

	return err
}

func (c *Consumer) monitor() {
	for {
		select {
		case <-c.done:
			return
		case err := <-c.closeErr:
			if err != nil {
				hlog.Errorf("Connection closed: %v, attempting to reconnect...", err)
				c.reconnect()
			}
		}
	}
}

func (c *Consumer) reconnect() {
	if c.reconnecting {
		return
	}
	c.reconnecting = true
	defer func() { c.reconnecting = false }()

	_ = c.client.conn.Close()
	_ = c.client.channel.Close()

	backoff := time.Second
	for {
		time.Sleep(backoff)

		if backoff < MaxReconnectDelay {
			backoff *= 2
		}
		if backoff > MaxReconnectDelay {
			backoff = MaxReconnectDelay
		}

		newClient, err := NewClient(c.client.address)
		if err != nil {
			hlog.Errorf("Failed to reconnect: %v", err)
			continue
		}

		c.client.conn = newClient.conn
		c.client.channel = newClient.channel

		c.client.conn.NotifyClose(c.closeErr)

		if err := c.consume(); err != nil {
			hlog.Errorf("Failed to start consuming: %v", err)
			continue
		}

		hlog.Info("Successfully reconnected and resumed consuming")
		break
	}
}

func (c *Producer) Publish(message []byte) error {
	return c.client.channel.Publish(
		c.config.Exchange,   // exchange
		c.config.RoutingKey, // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         message,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
}

func (c *Producer) PublishWithConfirm(message []byte) error {
	if err := c.client.channel.Confirm(false); err != nil {
		return fmt.Errorf("failed to set confirm mode: %v", err)
	}

	confirms := c.client.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	err := c.client.channel.Publish(
		c.config.Exchange,   // exchange
		c.config.RoutingKey, // routing key
		true,                // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         message,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return fmt.Errorf("message publish was not confirmed")
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("confirmation timeout")
	}

	return nil
}
