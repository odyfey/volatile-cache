package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"log"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/zullin/volatile-cache/server"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:20153")
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't conntect to server"))
	}

	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)

	msg := server.Message{
		Action:   "PUT",
		Key:      "test",
		Value:    "xpate",
		Lifetime: 20 * time.Second,
	}
	_ = enc.Encode(&msg)
	if err := write(conn, buf.Bytes()); err != nil {
		log.Println(errors.Wrapf(err, "can't sent message: %+v", msg))
	}
	buf.Reset()
	conn.Close()

	conn, err = net.Dial("tcp", "localhost:20153")
	defer conn.Close()
	if err != nil {
		log.Fatal(errors.Wrap(err, "can't conntect to server"))
	}
	time.Sleep(2 * time.Second)

	buf2 := new(bytes.Buffer)
	enc2 := gob.NewEncoder(buf2)
	msg2 := server.Message{
		Action: "READ",
		Key:    "test",
	}
	_ = enc2.Encode(&msg2)
	if err := write(conn, buf2.Bytes()); err != nil {
		log.Println(errors.Wrapf(err, "can't sent message: %+v", msg2))
	}
}

func write(conn net.Conn, content []byte) error {
	writer := bufio.NewWriter(conn)
	_, err := writer.Write(content)
	if err == nil {
		err = writer.Flush()
	}
	return err
}
