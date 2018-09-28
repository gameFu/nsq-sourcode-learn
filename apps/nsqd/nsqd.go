package main

import (
	"log"
	"nsq-learn/nsqd"
	"syscall"

	"github.com/judwhite/go-svc/svc"
)

type progarm struct {
	nsqd *nsqd.NSQD
}

func main() {
	prg := &progarm{}
	if err := svc.Run(prg, syscall.SIGINT, syscall.SIGTERM); err != nil {
		log.Fatal(err)
	}
}

func (p *progarm) Init(env svc.Environment) error {
	return nil
}

func (p *progarm) Start() error {
	opts := nsqd.NewOptions()
	nsqd := nsqd.New(opts)
	nsqd.Main()
	p.nsqd = nsqd
	return nil
}

func (p *progarm) Stop() error {
	return nil
}
