package nsqd

import (
	"nsq-learn/internal/lg"
	"nsq-learn/internal/util"
	"strings"
	"sync"
	"sync/atomic"

	diskqueue "github.com/nsqio/go-diskqueue"
)

type Topic struct {
	sync.RWMutex
	name      string
	startChan chan int
	exitChan  chan int
	waitGroup util.WaitGroupWrapper
	ctx       *context
	// 信息存储队列（用来持久化消息）
	backend   BackendQueue
	ephemeral bool
	// 是否暂停
	paused int32
}

func NewTopic(topicName string, ctx *context, deleteCallback func(*Topic)) *Topic {
	t := &Topic{
		name:      topicName,
		ctx:       ctx,
		startChan: make(chan int, 1),
		exitChan:  make(chan int),
	}
	// 如果topic名后面有#ephemeral则为临时topic
	if strings.HasSuffix(topicName, "#ephemeral") {
		t.ephemeral = true
		t.backend = newDummyBackendQueue()
	} else {
		dqLogf := func(level diskqueue.LogLevel, f string, args ...interface{}) {
			opts := ctx.nsqd.getOpts()
			lg.Logf(opts.Logger, opts.logLevel, lg.LogLevel(level), f, args...)
		}
		t.backend = diskqueue.New(
			topicName,
			ctx.nsqd.getOpts().DataPath,
			ctx.nsqd.getOpts().MaxBytesPerFile,
			int32(minValidMsgLength),
			int32(ctx.nsqd.getOpts().MaxMsgSize)+minValidMsgLength,
			ctx.nsqd.getOpts().SyncEvery,
			ctx.nsqd.getOpts().SyncTimeout,
			dqLogf,
		)
	}
	// 通知nsqd，进行持久化操作
	t.ctx.nsqd.Notify(t)
	return t
}

// 当前topic是否暂停
func (t *Topic) IsPaused() bool {
	return atomic.LoadInt32(&t.paused) == 1
}

// 开始Topic服务
func (t *Topic) Start() {

}
