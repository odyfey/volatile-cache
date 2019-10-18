package main

import (
	"bytes"
	"encoding/gob"
	"io"
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
	defer conn.Close()

	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)

	msg := server.Message{
		Action:   "PUT",
		Key:      "test",
		Value:    "xpate",
		Lifetime: 20 * time.Second,
	}
	_ = enc.Encode(&msg)
	if _, err := buf.WriteTo(conn); err != nil {
		log.Println(errors.Wrapf(err, "can't sent message: %+v", msg))
	}
	// buf.Reset()

	time.Sleep(2 * time.Second)
	buf2 := new(bytes.Buffer)
	enc2 := gob.NewEncoder(buf2)
	msg2 := server.Message{
		Action: "READ",
		Key:    "test",
	}
	_ = enc2.Encode(&msg2)
	if _, err := buf2.WriteTo(conn); err != nil {
		log.Println(errors.Wrapf(err, "can't sent message: %+v", msg2))
	}
	// buf2.Reset()

	time.Sleep(2 * time.Second)
	dbuf := new(bytes.Buffer)
	io.Copy(dbuf, conn)

	var resp server.Message
	dec := gob.NewDecoder(dbuf)
	_ = dec.Decode(&resp)
	log.Printf("response: %+v", resp)
}
