package worker

import "github.com/twmb/franz-go/pkg/kgo"

type ShardWork struct {
	Records []*kgo.Record
	Done    chan error
}
