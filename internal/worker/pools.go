package worker

import (
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

var recordsPool = sync.Pool{
	New: func() any {
		const batchSize = 50000
		const avgProtoSize = 49
		const safetyMargin = 15
		const totalByteSlabSize = batchSize * (avgProtoSize + safetyMargin)

		r := &RecordBatch{
			Slab:     make([]kgo.Record, batchSize),   // Preallocate slab of records for cache locality
			Pointers: make([]*kgo.Record, batchSize),  // Preallocate slice of pointers for Kafka client interface
			ByteSlab: make([]byte, totalByteSlabSize), // Preallocate byte slab for protobuf payloads
			offset:   0,
		}
		for i := range r.Slab {
			r.Pointers[i] = &r.Slab[i] // Point to slab records
		}
		return r
	},
}
