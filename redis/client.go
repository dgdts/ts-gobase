package redis

import (
	"context"
	"sync"
	"time"

	redis "github.com/redis/go-redis/v9"
)

const defaultClientName = "default"

type RedisClient struct {
	UniversalAddrs []string `yaml:"universal_addrs"`
	Password       string   `yaml:"password"`
	IdleTimeout    int      `yaml:"idle_timeout"`
	DB             int      `yaml:"db"`
	PoolSize       int      `yaml:"pool_size"`
	MasterName     string   `yaml:"master_name"`

	client redis.UniversalClient
	once   sync.Once
}

type redisClientManager struct {
	connectionMap map[string]*RedisClient
}

var redisClientManagerInstance *redisClientManager
var redisClientManagerInstanceOnce sync.Once

func getRedisClientManagerInstance() *redisClientManager {
	redisClientManagerInstanceOnce.Do(func() {
		redisClientManagerInstance = &redisClientManager{
			connectionMap: make(map[string]*RedisClient),
		}
	})
	return redisClientManagerInstance
}

func (rcm *redisClientManager) updateConfigs(configs map[string]*RedisClient) {
	rcm.connectionMap = configs
}

func (rcm *redisClientManager) getClient(name string) redis.UniversalClient {
	client, ok := rcm.connectionMap[name]
	if !ok {
		panic("cannot get redis name:" + name)
	}
	return client.connect()
}

func (r *RedisClient) connect() redis.UniversalClient {
	r.once.Do(func() {
		switch {
		case r.MasterName != "":
			r.client = redis.NewUniversalClient(&redis.UniversalOptions{
				Addrs:           r.UniversalAddrs,
				MasterName:      r.MasterName,
				Password:        r.Password,
				PoolSize:        r.PoolSize,
				ConnMaxIdleTime: time.Duration(r.IdleTimeout) * time.Second,
			})
		case len(r.UniversalAddrs) == 1:
			r.client = redis.NewUniversalClient(&redis.UniversalOptions{
				Addrs:           r.UniversalAddrs,
				Password:        r.Password,
				PoolSize:        r.PoolSize,
				ConnMaxIdleTime: time.Duration(r.IdleTimeout) * time.Second,
				DB:              r.DB,
			})
		default:
			r.client = redis.NewUniversalClient(&redis.UniversalOptions{
				Addrs:           r.UniversalAddrs,
				Password:        r.Password,
				PoolSize:        r.PoolSize,
				ConnMaxIdleTime: time.Duration(r.IdleTimeout) * time.Second,
			})
		}

		_, err := r.client.Ping(context.Background()).Result()
		if err != nil {
			panic("connect redis[" + r.UniversalAddrs[0] + "] failed:" + err.Error())
		}
	})
	return r.client
}

func RegisterConnection(configs map[string]*RedisClient) {
	getRedisClientManagerInstance().updateConfigs(configs)
}

func GetConnection(redisName ...string) redis.UniversalClient {
	var clientName string
	if len(redisName) == 0 {
		clientName = defaultClientName
	} else {
		clientName = redisName[0]
	}

	return getRedisClientManagerInstance().getClient(clientName)
}
