package store

import (
	"time"
)

type Watcher interface {
	Start(s Store)
	Stop(s Store)
}

type VolatileCacheWatcher struct {
	interval time.Duration
	stop     chan bool
}

func NewWatcher(i time.Duration) *VolatileCacheWatcher {
	return &VolatileCacheWatcher{
		interval: i,
		stop:     make(chan bool),
	}
}

func (w *VolatileCacheWatcher) Start(s Store) {
	tick := time.NewTicker(w.interval)
	for {
		select {
		case <-tick.C:
			s.DeleteExpired()
		case <-w.stop:
			tick.Stop()
			return
		}
	}
}

func (w *VolatileCacheWatcher) Stop(s Store) {
	w.stop <- true
}
