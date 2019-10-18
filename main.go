package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/zullin/volatile-cache/config"
	"github.com/zullin/volatile-cache/server"
)

func main() {
	cfg := config.GetInstance()
	srv := server.NewServer(cfg.Host, cfg.Port)

	r, err := os.Open(cfg.StoreFileName)
	if err != nil {
		log.Print(errors.Wrapf(err, "can't open file: %s", cfg.StoreFileName))
	}
	srv.FillServerCache(r)

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		w, err := os.Open(cfg.StoreFileName)
		if err == nil {
			srv.Stop(w)
		} else {
			log.Print(errors.Wrapf(err, "can't open file: %s", cfg.StoreFileName))
		}
		close(idleConnsClosed)
	}()
	srv.Start()
	<-idleConnsClosed
}
