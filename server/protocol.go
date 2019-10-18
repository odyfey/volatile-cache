package server

import (
	"time"
)

const (
	PutAction    = "PUT"
	ReadAction   = "READ"
	DeleteAction = "DELETE"
)

type Message struct {
	Action   string
	Key      string
	Value    string
	Lifetime time.Duration
}
