package nsqd

import (
	"log"
	"net"
	"nsq-learn/internal/dirlock"
	"nsq-learn/internal/http_api"
	"nsq-learn/internal/lg"
	"nsq-learn/internal/util"
	"nsq-learn/internal/version"
	"os"
	"sync/atomic"
	"time"
)

type NSQD struct {
	startTime    time.Time
	httpListener net.Listener
	// 暂不明白，为甚需要对opts做原子操作
	opts atomic.Value
	// 路径锁
	dl        *dirlock.DirLock
	waitGroup util.WaitGroupWrapper
}

func New(opts *Options) *NSQD {
	dataPath := opts.DataPath
	// 如果没有设置路劲，就将路径放在当前目录
	if dataPath == "" {
		cwd, _ := os.Getwd()
		dataPath = cwd
	}
	n := &NSQD{
		startTime: time.Now(),
		dl:        dirlock.New(dataPath),
	}
	// 初始化logger
	if opts.Logger == nil {
		opts.Logger = log.New(os.Stderr, opts.LogPrefix, log.Ldate|log.Ltime|log.Lmicroseconds)
	}
	// 将opts存入（首先将默认值存到原子值里）
	n.swapOpts(opts)
	// 将opts的LogLevel类型进行转换
	var err error
	opts.logLevel, err = lg.ParseLogLevel(opts.LogLevel, opts.Verbose)
	if err != nil {
		// lg.LogLevel是int的别名，所以默认是0，也就是DEBUG
		n.logf(LOG_FATAL, "%s", err)
		os.Exit(1)
	}
	// 锁定目录, 最简单的例子，如果再有nsqd启动目录设置为这个目录就会报错
	err = n.dl.Lock()
	if err != nil {
		n.logf(LOG_FATAL, "--data-path=%s in use (possibly by another instance of nsqd)", dataPath)
		os.Exit(1)
	}
	n.logf(LOG_INFO, version.String("nsqd"))
	n.logf(LOG_INFO, "ID: %d", opts.ID)
	return n
}

func (n *NSQD) Main() {
	var err error
	ctx := &context{n}
	n.httpListener, err = net.Listen("tcp", n.getOpts().HTTPAddress)
	if err != nil {
		n.logf(LOG_FATAL, "listen http (%s) failed - %s", n.getOpts().HTTPAddress, err)
		os.Exit(1)
	}
	// http server
	httpServer := NewHttpServer(ctx, false, false)
	// 异步启动
	n.waitGroup.Wrap(func() {
		http_api.Serve(n.httpListener, httpServer, "HTTP", n.logf)
	})
}

func (n *NSQD) swapOpts(opts *Options) {
	n.opts.Store(opts)
}

func (n *NSQD) getOpts() *Options {
	return n.opts.Load().(*Options)
}

// 获取nsqd进程的健康状况
func (n *NSQD) getHealth() string {
	return "OK"
}

// 判断nsqd是否健康
func (n *NSQD) isHealth() bool {
	return true
}
