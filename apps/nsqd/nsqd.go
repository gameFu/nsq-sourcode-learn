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
	// 导入元数据
	err := nsqd.LoadMetadata()
	if err != nil {
		// 等于打印 + 退出
		log.Fatalf("ERROR: %s", err.Error())
	}
	// 持久化元数据
	err = nsqd.PersistMetadata()
	if err != nil {
		log.Fatalf("ERROR: failed to persist metadata - %s", err.Error())
	}
	nsqd.Main()
	p.nsqd = nsqd
	return nil
}

// 结束
func (p *progarm) Stop() error {
	if p.nsqd != nil {
		p.nsqd.Exit()
	}
	return nil
}
