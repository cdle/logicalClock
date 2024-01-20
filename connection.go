package logical_clock

type Connection interface {
	Recv() (message *Message, ack func())
	Send(*Message) (clock Clock)
	Pid() int
}
