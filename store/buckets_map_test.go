package store

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucketSet(t *testing.T) {
	bm := NewBucketsMap(256)
	for _, item := range data {
		bm.Set(item.key, item.value, item.exp)
	}

}

func TestBucketSave(t *testing.T) {
	bm := NewBucketsMap(16)
	for _, item := range data {
		bm.Set(item.key, item.value, item.exp)
	}
	bm.pauseBucketsWatchers()

	var buf bytes.Buffer
	err := bm.Save(&buf)
	assert.Nil(t, err)
}

func TestBucketLoad(t *testing.T) {
	bm := NewBucketsMap(16)
	for _, item := range data {
		bm.Set(item.key, item.value, item.exp)
	}
	bm.pauseBucketsWatchers()

	var buf bytes.Buffer
	err := bm.Save(&buf)
	assert.Nil(t, err)

	bm2 := NewBucketsMap(16)
	err = bm2.Load(&buf)
	assert.Nil(t, err)
}
