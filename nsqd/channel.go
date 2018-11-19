package nsqd

import (
	"nsq-learn/internal/lg"
	"strings"
	"sync"

	diskqueue "github.com/nsqio/go-diskqueue"
)

type Channel struct {
	sync.RWMutex
	topicName      string
	name           string
	ctx            *context
	deleteCallback func(*Channel)
	// 是否为测试队列
	ephemeral bool
	// 持久化
	backend BackendQueue
}

// 创建一个新的channel
func NewChannel(topicName string, channelName string, ctx *context, deleteCallback func(*Channel)) *Channel {
	c := &Channel{
		topicName:      topicName,
		name:           channelName,
		ctx:            ctx,
		deleteCallback: deleteCallback,
	}

	//这里会设置优先队列和消息相关（先不处理）
	//持久化channel
	if strings.HasSuffix(channelName, "#ephemeral") {
		c.ephemeral = true
		c.backend = newDummyBackendQueue()
	} else {
		dqLogf := func(level diskqueue.LogLevel, f string, args ...interface{}) {
			opts := ctx.nsqd.getOpts()
			lg.Logf(opts.Logger, opts.logLevel, lg.LogLevel(level), f, args...)
		}
		// 每个topic都唯一
		backendName := getBackendName(topicName, channelName)
		c.backend = diskqueue.New(
			backendName,
			ctx.nsqd.getOpts().DataPath,
			ctx.nsqd.getOpts().MaxBytesPerFile,
			int32(minValidMsgLength),
			int32(ctx.nsqd.getOpts().MaxMsgSize)+minValidMsgLength,
			ctx.nsqd.getOpts().SyncEvery,
			ctx.nsqd.getOpts().SyncTimeout,
			dqLogf,
		)
	}
	// 持久化channel
	return c
}
