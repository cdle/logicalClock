package logical_clock

const (
	MessageTypeLock = iota
	MessageTypeUnLock
	MessageTypeAck
	MessageTypePrepare
)

type Message struct {
	Pid     int
	Type    int
	Clock   *Clock
	Carrier chan *Clock
}
