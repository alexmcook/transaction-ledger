package api

import (
	"errors"

	"github.com/twmb/franz-go/pkg/kgo"
)

type RecordBatch struct {
	Slab     []kgo.Record
	Pointers []*kgo.Record
	ByteSlab []byte
	offset   int
}

func (r *RecordBatch) Reset(count int) {
	r.offset = 0
	for i := range count {
		r.Slab[i].Value = nil
		r.Slab[i].Key = nil
	}
}

func (r *RecordBatch) NextRecord(size int) ([]byte, error) {
	if r.offset+size > len(r.ByteSlab) {
		return nil, errors.New("byte slab exhausted")
	}
	buf := r.ByteSlab[r.offset : r.offset+size]
	r.offset += size
	return buf, nil
}
