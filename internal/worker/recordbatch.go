package worker

import (
	"errors"

	"github.com/twmb/franz-go/pkg/kgo"
)

type RecordBatch struct {
	Slab     []kgo.Record
	Pointers []*kgo.Record
	ByteSlab []byte
	Count    int
	offset   int
}

func (r *RecordBatch) Reset() {
	r.offset = 0
	r.Count = 0
}

func (r *RecordBatch) NextRecord(size int) ([]byte, error) {
	if r.offset+size > len(r.ByteSlab) {
		return nil, errors.New("byte slab exhausted")
	}
	buf := r.ByteSlab[r.offset : r.offset+size]
	r.offset += size
	return buf, nil
}
