package nsqd

import (
	"crypto/md5"
	"hash/crc32"
	"io"
	"log"
	"nsq-learn/internal/lg"
	"os"
)

type Options struct {
	// 每一个nsqd都会有一个独立的id,为以后做分布式做准备
	ID          int64
	LogLevel    string
	LogPrefix   string
	HTTPAddress string
	// 存放数据的路径
	DataPath string

	logLevel lg.LogLevel //私有的，原因是需要转换成lg.LogLevel
	Logger   Logger
	Verbose  bool //官方说为了向后兼容，先不管
}

func NewOptions() *Options {
	defaultID := generateDefaultID()
	return &Options{
		ID:          defaultID,
		LogPrefix:   "[nsqd] ",
		LogLevel:    "info",
		Verbose:     false,
		HTTPAddress: "0.0.0.0:1418",
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
