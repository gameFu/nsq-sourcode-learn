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
	// channel表
	channelMap map[string]*Channel
}

func NewTopic(topicName string, ctx *context, deleteCallback func(*Topic)) *Topic {
	t := &Topic{
		name:       topicName,
		ctx:        ctx,
		startChan:  make(chan int, 1),
		exitChan:   make(chan int),
		channelMap: make(map[string]*Channel),
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

// 查找或创建channel(线程安全)
func (t *Topic) GetChannel(channelName string) *Channel {
	t.Lock()
	channel, isNew := t.getOrCreateChannel(channelName)
	t.Unlock()
	// 如果是新的channel，需要触发更新channel通知(待做)
	if isNew {
	}
	return channel
}

func (t *Topic) getOrCreateChannel(channelName string) (*Channel, bool) {
	channel, ok := t.channelMap[channelName]
	if !ok {
		deleteCallback := func(c *Channel) {
			t.DeleteExistingChannel(c.name)
		}
		channel := NewChannel(t.name, channelName, t.ctx, deleteCallback)
		t.channelMap[channelName] = channel
		t.ctx.nsqd.logf(LOG_INFO, "TOPIC(%s): new channel(%s)", t.name, channel.name)
		return channel, true
	}
	return channel, false
}

// 删除已经存在的channel，需要，移除channel表中的数据，执行channel的删除操作，触发更新channel信息通知等（待做）
func (t *Topic) DeleteExistingChannel(channelName string) error {
	return nil
}

// 当前topic是否暂停
func (t *Topic) IsPaused() bool {
	return atomic.LoadInt32(&t.paused) == 1
}

// 开始Topic服务
func (t *Topic) Start() {

}

// 关闭Topic
func (t *Topic) Close() {
	// 关闭文件系统
	t.backend.Close()
}
