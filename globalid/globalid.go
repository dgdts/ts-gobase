package globalid

import (
	"github.com/bwmarrin/snowflake"
)

type Option func(*option)

type option struct {
	machineID int64
}

func WithMachineID(machineID int64) Option {
	return func(o *option) {
		o.machineID = machineID
	}
}

type IDGenerator struct {
	node *snowflake.Node
	opts *option
}

func New(opts ...Option) (*IDGenerator, error) {
	ret := &IDGenerator{
		opts: &option{},
	}

	for _, opt := range opts {
		opt(ret.opts)
	}

	var err error
	ret.node, err = snowflake.NewNode(ret.opts.machineID)

	return ret, err
}

func (g *IDGenerator) Generate64() int64 {
	return g.node.Generate().Int64()
}

func (g *IDGenerator) GenerateBase36() string {
	return g.node.Generate().Base36()
}
