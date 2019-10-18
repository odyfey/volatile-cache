package store

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var data = []struct {
	key   string
	value string
	exp   time.Duration
}{
	{
		"hello",
		"world",
		5 * time.Second,
	},
	{
		"name",
		"qwerty",
		3 * time.Millisecond,
	},
	{
		"123",
		"987",
		10 * time.Second,
	},
}

func TestSet(t *testing.T) {
	cm := newConcurrentTTLMap(3 * time.Second)
	for _, item := range data {
		cm.Set(item.key, item.value, item.exp)
	}
	assert.Equal(t, len(data), len(cm.items))
}

func TestGet(t *testing.T) {
	cm := prepareMap()

	_, ok := cm.Get("notFoundKey")
	assert.False(t, ok)

	_, ok = cm.Get("123")
	assert.True(t, ok)
}

func TestDelete(t *testing.T) {
	cm := prepareMap()
	cm.Delete("hello")
	cm.Delete("fakeKey")
}

func TestDeleteExpired(t *testing.T) {
	var data = []struct {
		key   string
		value string
		exp   time.Duration
	}{
		{
			"hello",
			"world",
			150 * time.Millisecond,
		},
		{
			"name",
			"qwerty",
			30 * time.Millisecond,
		},
	}

	cm := newConcurrentTTLMap(100 * time.Millisecond)
	for _, item := range data {
		cm.Set(item.key, item.value, item.exp)
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 1, len(cm.items))
}

func TestSave(t *testing.T) {
	cm := prepareMap()
	cm.pauseWatcher()

	var buf bytes.Buffer
	err := cm.Save(&buf)
	assert.Nil(t, err)
}

func TestLoadToCurrent(t *testing.T) {
	cm := prepareMap()
	cm.pauseWatcher()

	var buf bytes.Buffer
	err := cm.Save(&buf)
	assert.Nil(t, err)

	err = cm.Load(&buf)
	assert.Nil(t, err)

	assert.Equal(t, len(data), len(cm.items))
}

func TestLoadToAnother(t *testing.T) {
	cm := prepareMap()
	cm.pauseWatcher()

	var buf bytes.Buffer
	err := cm.Save(&buf)
	assert.Nil(t, err)

	cm2 := newConcurrentTTLMap(3 * time.Second)
	err = cm2.Load(&buf)
	assert.Nil(t, err)
	assert.Equal(t, len(data), len(cm2.items))
}

func prepareMap() *concurrentTTLMap {
	cm := newConcurrentTTLMap(3 * time.Second)
	for _, item := range data {
		cm.Set(item.key, item.value, item.exp)
	}
	return cm
}
