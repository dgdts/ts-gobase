package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/segmentio/kafka-go"
)

type Client struct {
	brokers []string
}

type Producer struct {
	client *Client
	config *ProducerConf
	writer *kafka.Writer
}

type Consumer struct {
	client  *Client
	config  *ConsumerConf
	reader  *kafka.Reader
	handler func([]byte) error
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{}
}

func NewClient(brokers []string) (*Client, error) {
	return &Client{
		brokers: brokers,
	}, nil
}

func InitProducer(conf *ProducerConf) (*Producer, error) {
	client, err := NewClient(conf.Brokers)
	if err != nil {
		return nil, err
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(conf.Brokers...),
		Topic:        conf.Topic,
		Async:        conf.Async,
		RequiredAcks: kafka.RequireAll,
		MaxAttempts:  DefaultRetryCount,
	}

	return &Producer{
		client: client,
		config: conf,
		writer: writer,
	}, nil
}

func InitConsumer(conf *ConsumerConf) (*Consumer, error) {
	client, err := NewClient(conf.Brokers)
	if err != nil {
		return nil, err
	}

	if conf.RetryCount <= 0 {
		conf.RetryCount = DefaultRetryCount
	}

	config := kafka.ReaderConfig{
		Brokers:     conf.Brokers,
		Topic:       conf.Topic,
		GroupID:     conf.Group,
		MaxAttempts: conf.RetryCount,
	}

	if conf.FromBeginning {
		config.StartOffset = kafka.FirstOffset
	}

	reader := kafka.NewReader(config)

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		client: client,
		config: conf,
		reader: reader,
		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
	}, nil
}

func (p *Producer) Publish(message []byte) error {
	msg := kafka.Message{
		Value: message,
	}
	return p.writer.WriteMessages(context.Background(), msg)
}

func (c *Consumer) Start(handler func([]byte) error) error {
	c.handler = handler

	go func() {
		for {
			select {
			case <-c.done:
				return
			default:
				msg, err := c.reader.ReadMessage(c.ctx)
				if err != nil {
					if c.ctx.Err() == context.Canceled {
						return
					}
					hlog.Errorf("Error reading message: %v", err)
					continue
				}

				if err := c.processMessage(msg); err != nil {
					hlog.Errorf("Error processing message: %v", err)
				}
			}
		}
	}()

	return nil
}

func (c *Consumer) Stop() {
	c.cancel()
	close(c.done)
	if err := c.reader.Close(); err != nil {
		hlog.Errorf("Error closing consumer: %v", err)
	}
}

func (c *Consumer) processMessage(msg kafka.Message) error {
	var err error
	for attempt := 0; attempt <= c.config.RetryCount; attempt++ {
		err = c.handler(msg.Value)
		if err == nil {
			if !c.config.AutoCommit {
				if err = c.reader.CommitMessages(c.ctx, msg); err != nil {
					return fmt.Errorf("failed to commit message: %v", err)
				}
			}
			return nil
		}

		hlog.Warnf("Processing attempt %d failed: %v", attempt+1, err)

		if attempt < c.config.RetryCount {
			time.Sleep(time.Duration(c.config.RetryDelay) * time.Second)
		}
	}

	return err
}
