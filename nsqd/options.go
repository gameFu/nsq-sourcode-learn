package nsqd

import (
	"crypto/md5"
	"hash/crc32"
	"io"
	"log"
	"nsq-learn/internal/lg"
	"os"
	"time"
)

type Options struct {
	// 每一个nsqd都会有一个独立的id,为以后做分布式做准备
	ID          int64
	LogLevel    string
	LogPrefix   string
	HTTPAddress string
	// 存放数据的路径
	DataPath string

	logLevel        lg.LogLevel //私有的，原因是需要转换成lg.LogLevel
	Logger          Logger
	Verbose         bool          //官方说为了向后兼容，先不管
	MaxBytesPerFile int64         //当个文件最大容量（用来持久化消息）
	MaxMsgSize      int64         //消息最大的尺寸
	SyncEvery       int64         //暂时不明
	SyncTimeout     time.Duration //持久化，同步超时时间
}

func NewOptions() *Options {
	defaultID := generateDefaultID()
	return &Options{
		ID:              defaultID,
		LogPrefix:       "[nsqd] ",
		LogLevel:        "info",
		Verbose:         false,
		HTTPAddress:     "0.0.0.0:1418",
		MaxBytesPerFile: 100 * 1024 * 1024,
		SyncEvery:       2500,
		SyncTimeout:     2 * time.Second,
	}
}

func generateDefaultID() int64 {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	h := md5.New()
	io.WriteString(h, hostname)
	return int64(crc32.ChecksumIEEE(h.Sum(nil)) % 1024)
}
