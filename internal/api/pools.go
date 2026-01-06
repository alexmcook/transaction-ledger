package api

import (
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

var recordsPool = sync.Pool{
	New: func() any {
		const batchSize = 1000
		const avgProtoSize = 49
		const safetyMargin = 15
		const totalByteSlabSize = batchSize * (avgProtoSize + safetyMargin)

		r := &RecordBatch{
			Slab:     make([]kgo.Record, 1000),        // Preallocate slab of records for cache locality
			Pointers: make([]*kgo.Record, 1000),       // Preallocate slice of pointers for Kafka client interface
			ByteSlab: make([]byte, totalByteSlabSize), // Preallocate byte slab for protobuf payloads
			offset:   0,
		}
		for i := range r.Slab {
			r.Pointers[i] = &r.Slab[i] // Point to slab records
		}
		return r
	},
}

var trPool = sync.Pool{
	New: func() any {
		j := make([]TransactionRequest, 1000) // Preallocate for batch size of 1000
		return &j
	},
}
