package store

import (
	"encoding/gob"
	"hash/crc32"
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/zullin/volatile-cache/config"
)

type BucketsMap struct {
	buckets    []*concurrentTTLMap
	bucketsNum int
}

func NewBucketsMap(bucketsNum int) *BucketsMap {
	cfg := config.GetInstance()
	bm := &BucketsMap{
		buckets:    make([]*concurrentTTLMap, bucketsNum),
		bucketsNum: bucketsNum,
	}
	for idx := range bm.buckets {
		bm.buckets[idx] = newConcurrentTTLMap(time.Duration(cfg.ExpirationTime) * time.Second)
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
	enc := gob.NewEncoder(w)
	for idx := range bm.buckets {
		if bm.buckets[idx].len() > 0 {
			if err := bm.buckets[idx].saveToStream(enc); err != nil {
				return errors.Wrapf(err, "error while saving bucket, index: %d", idx)
			}
		}
	}
	return nil
}

func (bm *BucketsMap) Load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	for {
		res := map[string]Item{}
		err := dec.Decode(&res)
		if err == io.EOF {
			break
		} else if err != nil {
			return errors.Wrapf(err, "error while loading buckets")
		}
		for key, item := range res {
			idx := bm.calculateBucketIndex(key)
			bm.buckets[idx].setPreparedItem(key, item)
		}
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
