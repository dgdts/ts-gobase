package mongo

import (
	"time"

	"github.com/kamva/mgm/v3"
	mongoOption "go.mongodb.org/mongo-driver/mongo/options"
)

type OrmClient struct {
	MongoClientConfig
	opt *mgm.Config
}

func WithTimeout(timeout time.Duration) func(opt *mgm.Config) {
	return func(opt *mgm.Config) {
		opt.CtxTimeout = timeout
	}
}

var ormClient *OrmClient

func InitMongoDB(configs *MongoClientConfig, opts ...func(opt *mgm.Config)) error {
	ormClient = &OrmClient{
		MongoClientConfig: *configs,
	}

	for _, opt := range opts {
		if ormClient.opt == nil {
			ormClient.opt = &mgm.Config{}
		}

		opt(ormClient.opt)
	}

	err := mgm.SetDefaultConfig(
		ormClient.opt,
		ormClient.Database,
		mongoOption.Client().ApplyURI(ormClient.getURI()),
	)

	return err
}
