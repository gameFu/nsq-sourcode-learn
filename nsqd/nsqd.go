package nsqd

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"nsq-learn/internal/dirlock"
	"nsq-learn/internal/http_api"
	"nsq-learn/internal/lg"
	"nsq-learn/internal/util"
	"nsq-learn/internal/version"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"
)

type NSQD struct {
	startTime    time.Time
	httpListener net.Listener
	// 配置项
	opts atomic.Value
	// 路径锁
	dl        *dirlock.DirLock
	waitGroup util.WaitGroupWrapper
	// topic表
	topicMap map[string]*Topic
	// 是否在load metadata, 使用int32的原因是为了方便做原子操作
	isLoading int32
	// 退出chan
	exitChan   chan int
	notifyChan chan interface{}
	sync.RWMutex
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
		topicMap:  make(map[string]*Topic),
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

// 获取topic，如果没有就创建(线程安全)
func (n *NSQD) GetTopic(topicName string) *Topic {
	// 最好的情况，是已经存在，那么添加读锁
	n.RLock()
	t, ok := n.topicMap[topicName]
	n.RUnlock()
	if ok {
		return t
	}
	// 如果没有就创建一个新的topic，添加写锁
	n.Lock()
	// 防止并发的情况下，上一个写锁已经成功写入
	t, ok = n.topicMap[topicName]
	if ok {
		n.Lock()
		return t
	}
	// 暂时为无效的
	deleteCallback := func(t *Topic) {

	}
	// 创建一个新的topic
	t = NewTopic(topicName, &context{n}, deleteCallback)
	// 写到nsqd topic表中
	n.topicMap[topicName] = t
	n.Unlock()
	n.logf(LOG_INFO, "TOPIC(%s): created", t.name)

	// 如果正在loading则就不在继续，暂不明意思，待后续处理
	if atomic.LoadInt32(&n.isLoading) == 1 {
		return t
	}
	// TODO: 这里有lookup相关逻辑，先不处理
	t.Start()
	return t
}

// 触发这个方法，将会持久化metadata(包括channel和topic等数据)
func (n *NSQD) Notify(v interface{}) {
	// 判断是否处于loading状态，如果处于loading状态，那么，不用该进行presist metadata
	persist := atomic.LoadInt32(&n.isLoading) == 0
	n.waitGroup.Wrap(func() {
		select {
		case <-n.exitChan:
		// 这个通知与nsqdlookup相关,由于暂时不涉及lookup相关，先使用default
		// case n.notifyChan <- v:
		default:
			if !persist {
				return
			}
			n.Lock()
			// 保存元数据
			err := n.PersistMetadata()
			if err != nil {
				n.logf(LOG_ERROR, "failed to persist metadata - %s", err)
			}
			n.Unlock()
		}
	})
}

// 持久化topic和channel等信息，以便重启后能恢复数据
func (n *NSQD) PersistMetadata() error {
	fileName := newMetadataFile(n.getOpts())
	n.logf(LOG_INFO, "NSQ: persisting topic/channel metadata to %s", fileName)
	js := make(map[string]interface{})
	// 初始化一个空接口数组，用来存放topic相关数据
	topics := []interface{}{}
	// 遍历所有已经注册的topic
	for _, topic := range n.topicMap {
		// 如果为假的（测试用）topic， 则跳过持久化
		if topic.ephemeral {
			continue
		}
		topicData := make(map[string]interface{})
		topicData["name"] = topic.name
		topicData["paused"] = topic.IsPaused()
		// TODO:这里有channel相关持久化逻辑，后续再处理
		topics = append(topics, topicData)
	}
	js["version"] = version.Binary
	js["topics"] = topics
	// 进行json序列化
	data, err := json.Marshal(&js)
	if err != nil {
		return err
	}

	//为什么要这么做，这是为了安全的更新nsqd.dat文件，直接更新是不安全的，因为有可能有多个线程去更新这个文件，临时文件改名后，会吧原来的文件删除，并且吧内容替换为临时文件的内容
	//先创建一个临时文件
	tmpFileName := fmt.Sprintf("%s.%d.tmp", fileName, rand.Int())
	err = writeSyncFile(tmpFileName, data)
	// 更名
	err = os.Rename(tmpFileName, fileName)
	if err != nil {
		return err
	}
	return nil
}

func writeSyncFile(fn string, data []byte) error {
	// 只写，如果不存在就创建，清空
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	// 将内容写到文件
	_, err = f.Write(data)
	if err == nil {
		// sync是fsync系统调用，会将数据和元数据都刷到磁盘， 为什么要这么做，应为f.Write调用成功不保证，已经写到磁盘上了，因为操作系统会缓存操作
		err = f.Sync()
	}
	f.Close()
	return err
}

// 返回存储metadata file的文件路劲，存在当前datapath目录下的
func newMetadataFile(opts *Options) string {
	return path.Join(opts.DataPath, "nsqd.dat")
}
