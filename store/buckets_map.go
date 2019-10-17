package store

import (
	"hash/crc32"
	"time"
)

type BucketsMap struct {
	buckets    []*ConcurrentTTLMap
	bucketsNum int
}

func NewBucketsMap(bucketsNum int) *BucketsMap {
	bm := &BucketsMap{
		buckets:    make([]*ConcurrentTTLMap, bucketsNum),
		bucketsNum: bucketsNum,
	}
	for b := range bm.buckets {
		bm.buckets[b] = NewConcurrentMap(60 * time.Second) // todo: env config
	}
	return bm
}

func (bm *BucketsMap) calculateBucketIndex(key string) uint32 {
	hash := crc32.ChecksumIEEE([]byte(key))
	return hash % uint32(bm.bucketsNum)
}

func (bm *BucketsMap) Set(key, value string, exp time.Duration) {
	ind := bm.calculateBucketIndex(key)
	bm.buckets[ind].Set(key, value, exp)
}

func (bm *BucketsMap) Get(key string) (string, bool) {
	ind := bm.calculateBucketIndex(key)
	return bm.buckets[ind].Get(key)
}

func (bm *BucketsMap) Delete(key string) {
	ind := bm.calculateBucketIndex(key)
	bm.buckets[ind].Delete(key)
}
