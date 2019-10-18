package server

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"

	"github.com/pkg/errors"
	"github.com/zullin/volatile-cache/store"
)

type Server struct {
	host  string
	port  int
	cache store.Store
}

func NewServer(host string, port int) *Server {
	return &Server{
		host:  host,
		port:  port,
		cache: store.NewBucketsMap(256),
	}
}

func (s *Server) Start(host string, port int) error {
	address := fmt.Sprintf("%s:%d", host, port)
	l, err := net.Listen("tcp", address)
	if err != nil {
		return errors.Wrapf(err, "can't start tcp listener on %s", address)
	}
	defer l.Close()
	log.Printf("Listening on: %s", address)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(errors.Wrap(err, "can't accept message"))
		}
		go s.handleRequest(conn)
	}
}

func (s *Server) handleRequest(conn net.Conn) {
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		log.Println(errors.Wrap(err, "can't read message"))
		return
	}
	var msg Message
	dec := gob.NewDecoder(bytes.NewBuffer(buf))
	if err = dec.Decode(&msg); err != nil {
		log.Println(errors.Wrap(err, "can't decode message"))
	}

	res, err := s.processAction(msg)
	if err != nil {
		log.Println(errors.Wrapf(err, "can't run action %s", msg.Action))
	}
	if _, err = conn.Write([]byte(res)); err != nil {
		log.Println(errors.Wrap(err, "can't sent response"))
	}
}

func (s *Server) processAction(msg Message) (res string, err error) {
	switch msg.Action {
	case PutAction:
		s.cache.Set(msg.Key, msg.Value, msg.Lifetime)

	case ReadAction:
		res, _ = s.cache.Get(msg.Key)

	case DeleteAction:
		s.cache.Delete(msg.Key)

	default:
		err = errors.Errorf("unexpected message action: %s", msg.Action)
	}
	return
}
