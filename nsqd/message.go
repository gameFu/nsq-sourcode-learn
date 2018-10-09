package nsqd

const (
	MsgIDLength = 16
	// 最小的消息合法长度
	minValidMsgLength = MsgIDLength + 8 + 2 // Timestamp + Attempts
)
