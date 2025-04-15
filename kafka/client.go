package kafka

import (
	"time"
)

const (
	DefaultRetryCount    = 3
	DefaultRetryDelay    = 2 * time.Second
	DefaultBatchSize     = 100
	DefaultCommitTimeout = 5 * time.Second
)

var (
	initByKeysEnable = false
	initKeys         = make([]string, 0)
	producers        = make(map[string]*Producer)
	consumers        = make(map[string]*Consumer)
)

// Kafka 配置结构
type Kafka struct {
	Brokers  []string        `yaml:"brokers"` // Kafka broker 地址列表
	Producer []*ProducerConf `yaml:"producers"`
	Consumer []*ConsumerConf `yaml:"consumers"`
}

type ProducerConf struct {
	Key     string   `yaml:"key"`     // 生产者标识
	Brokers []string `yaml:"brokers"` // broker地址列表
	Topic   string   `yaml:"topic"`   // 主题
	Async   bool     `yaml:"async"`   // 是否异步发送
}

type ConsumerConf struct {
	Key           string   `yaml:"key"`            // 消费者标识
	Brokers       []string `yaml:"brokers"`        // broker地址列表
	Topic         string   `yaml:"topic"`          // 主题
	Group         string   `yaml:"group"`          // 消费者组
	RetryCount    int      `yaml:"retry_count"`    // 重试次数
	RetryDelay    int      `yaml:"retry_delay"`    // 重试延迟（秒）
	AutoCommit    bool     `yaml:"auto_commit"`    // 自动提交
	FromBeginning bool     `yaml:"from_beginning"` // 是否从头开始消费
}

func RegisterConnWithKeys(k *Kafka, keys ...string) error {
	initKeys = keys
	initByKeysEnable = true
	return RegisterConn(k)
}

func RegisterConn(k *Kafka) error {
	producerConfs := k.Producer
	consumerConfs := k.Consumer

	for i, v := range producerConfs {
		if len(v.Brokers) == 0 {
			producerConfs[i].Brokers = k.Brokers
		}
	}
	for i, v := range consumerConfs {
		if len(v.Brokers) == 0 {
			consumerConfs[i].Brokers = k.Brokers
		}
	}

	filterByInitKeys(k)

	for _, conf := range producerConfs {
		producer, err := InitProducer(conf)
		if err != nil {
			return err
		}
		producers[conf.Key] = producer
	}

	for _, conf := range consumerConfs {
		consumer, err := InitConsumer(conf)
		if err != nil {
			return err
		}
		consumers[conf.Key] = consumer
	}

	return nil
}

func filterByInitKeys(kafka *Kafka) {
	if !initByKeysEnable {
		return
	}

	producerTmp := make([]*ProducerConf, 0)
	for _, v := range kafka.Producer {
		for _, key := range initKeys {
			if v.Key == key {
				producerTmp = append(producerTmp, v)
			}
		}
	}
	kafka.Producer = producerTmp

	consumerTmp := make([]*ConsumerConf, 0)
	for _, v := range kafka.Consumer {
		for _, key := range initKeys {
			if v.Key == key {
				consumerTmp = append(consumerTmp, v)
			}
		}
	}
	kafka.Consumer = consumerTmp
}

func GetProducer(key string) *Producer {
	return producers[key]
}

func GetConsumer(key string) *Consumer {
	return consumers[key]
}
