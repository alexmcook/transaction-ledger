package db

type BucketProvider interface {
	GetActiveBucket() int32
}
