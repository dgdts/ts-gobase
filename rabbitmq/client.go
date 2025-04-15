package rabbitmq

import (
	"time"
)

const (
	DefaultRetryCount     = 3
	DefaultRetryDelay     = 2 * time.Second
	DefaultPrefetchCount  = 1
	DefaultReconnectDelay = 5 * time.Second
	MaxReconnectDelay     = 300 * time.Second
	DefaultConfirmTimeout = 5 * time.Second
)

var (
	initByKeysEnable = false
	initKeys         = make([]string, 0)
	producers        = make(map[string]*Producer)
	consumers        = make(map[string]*Consumer)
)

// RabbitMQ 远程配置中心进行配置
type RabbitMQ struct {
	Address  string          `yaml:"address"` // 连接地址，格式：amqp://user:pass@host:port/vhost
	Producer []*ProducerConf `yaml:"producers"`
	Consumer []*ConsumerConf `yaml:"consumers"`
}

type ProducerConf struct {
	Key        string `yaml:"key"`         // 生产者标识
	Address    string `yaml:"address"`     // 连接地址，格式：amqp://user:pass@host:port/vhost
	Exchange   string `yaml:"exchange"`    // 交换机名称
	RoutingKey string `yaml:"routing_key"` // 路由键
	Confirm    bool   `yaml:"confirm"`     // 是否启用发布确认
}

type ConsumerConf struct {
	Key           string `yaml:"key"`            // 消费者标识
	Address       string `yaml:"address"`        // 连接地址，格式：amqp://user:pass@host:port/vhost
	Queue         string `yaml:"queue"`          // 队列名称
	PrefetchCount int    `yaml:"prefetch_count"` // 预取计数
	RetryCount    int    `yaml:"retry_count"`    // 重试次数
	RetryDelay    int    `yaml:"retry_delay"`    // 重试延迟（秒）
	AutoAck       bool   `yaml:"auto_ack"`       // 自动确认
}

func RegisterConnWithKeys(mq *RabbitMQ, keys ...string) error {
	initKeys = keys
	initByKeysEnable = true
	return RegisterConn(mq)
}

func RegisterConn(mq *RabbitMQ) error {
	producerConfs := mq.Producer
	consumerConfs := mq.Consumer

	for i, v := range producerConfs {
		if v.Address == "" {
			producerConfs[i].Address = mq.Address
		}
	}
	for i, v := range consumerConfs {
		if v.Address == "" {
			consumerConfs[i].Address = mq.Address
		}
	}

	filterByInitKeys(mq)

	for _, conf := range producerConfs {
		producer, err := InitProducer(conf)
		if err != nil {
			return err
		}
		producers[conf.Key] = producer
	}

	// 注册消费者
	for _, conf := range consumerConfs {
		consumer, err := InitConsumer(conf)
		if err != nil {
			return err
		}
		consumers[conf.Key] = consumer
	}

	return nil
}

func filterByInitKeys(rabbitMQ *RabbitMQ) {
	if !initByKeysEnable {
		return
	}

	producerTmp := make([]*ProducerConf, 0)
	for _, v := range rabbitMQ.Producer {
		for _, key := range initKeys {
			if v.Key == key {
				producerTmp = append(producerTmp, v)
			}
		}
	}
	rabbitMQ.Producer = producerTmp

	consumerTmp := make([]*ConsumerConf, 0)
	for _, v := range rabbitMQ.Consumer {
		for _, key := range initKeys {
			if v.Key == key {
				consumerTmp = append(consumerTmp, v)
			}
		}
	}
	rabbitMQ.Consumer = consumerTmp
}

func GetProducer(key string) *Producer {
	return producers[key]
}

func GetConsumer(key string) *Consumer {
	return consumers[key]
}
