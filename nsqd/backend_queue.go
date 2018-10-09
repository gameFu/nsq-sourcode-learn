package nsqd

// 消息持久化队列接口
type BackendQueue interface {
	Put([]byte) error
	ReadChan() chan []byte //预期是一个无缓冲的通道
	Close() error
	Delete() error
	Depth() int64
	Empty() error
}
