package store

import (
	"hash/crc32"
	"io"
	"time"

	"github.com/pkg/errors"
)

type BucketsMap struct {
	buckets    []*concurrentTTLMap
	bucketsNum int
}

func NewBucketsMap(bucketsNum int) *BucketsMap {
	bm := &BucketsMap{
		buckets:    make([]*concurrentTTLMap, bucketsNum),
		bucketsNum: bucketsNum,
	}
	for idx := range bm.buckets {
		bm.buckets[idx] = newConcurrentTTLMap(60 * time.Second) // todo: env config
	}
	return bm
}

func (bm *BucketsMap) Set(key, value string, exp time.Duration) {
	idx := bm.calculateBucketIndex(key)
	bm.buckets[idx].Set(key, value, exp)
}

func (bm *BucketsMap) Get(key string) (string, bool) {
	idx := bm.calculateBucketIndex(key)
	return bm.buckets[idx].Get(key)
}

func (bm *BucketsMap) Delete(key string) {
	idx := bm.calculateBucketIndex(key)
	bm.buckets[idx].Delete(key)
}

func (bm *BucketsMap) Save(w io.Writer) error {
	// todo: bug, сначала собрать все item, потом делать Save
	for idx := range bm.buckets {
		if bm.buckets[idx].len() > 0 {
			if err := bm.buckets[idx].Save(w); err != nil {
				return errors.Wrapf(err, "error while saving bucket, index: %d", idx)
			}
		}
	}
	return nil
}

// load all items in one map, then insert them into different buckets
func (bm *BucketsMap) Load(r io.Reader) error {
	c := newConcurrentTTLMap(60 * time.Second)
	if err := c.Load(r); err != nil {
		return errors.Wrap(err, "error while loading buckets")
	}

	for key, item := range c.itemsCopy() {
		idx := bm.calculateBucketIndex(key)
		bm.buckets[idx].setPreparedItem(key, item)
	}
	return nil
}

func (bm *BucketsMap) calculateBucketIndex(key string) uint32 {
	hash := crc32.ChecksumIEEE([]byte(key))
	return hash % uint32(bm.bucketsNum)
}

func (bm *BucketsMap) pauseBucketsWatchers() {
	for idx := range bm.buckets {
		bm.buckets[idx].pauseWatcher()
	}
}
