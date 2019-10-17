package store

import (
	"io"
	"time"
)

type Store interface {
	Set(key, value string, exp time.Duration)
	Get(key string) (string, bool)
	Delete(key string)
	Save(w io.Writer) error
	Load(r io.Reader) error
}
