package logical_clock

import (
	"fmt"
	"sync"
	"testing"
)

type MockConn struct {
	client *Client
	ch     chan *Message
}

var acks sync.Map

func (conn *MockConn) Pid() int {
	return conn.client.Clock.Pid
}

func (conn *MockConn) Recv() (message *Message, ack func()) {
	return <-conn.ch, func() {
		v, _ := acks.Load(message)
		v.(chan Clock) <- *conn.client.Clock
	}
}

func (conn *MockConn) Send(message *Message) Clock {
	ch := make(chan Clock)
	acks.Store(message, ch)
	conn.ch <- message
	return <-ch
}

func TestLlock(t *testing.T) {
	var conns = []*MockConn{}
	var peers = []Connection{}
	var clients = []*Client{}
	for i := 0; i < 3; i++ {
		var conn = &MockConn{ch: make(chan *Message)}
		conns = append(conns, conn)
		peers = append(peers, conn)
	}
	for i, conn := range conns {
		var client = new(Client)
		client.Clock = &Clock{Pid: i}
		conn.client = client
		client.Local = conn
		client.Peers = peers
		clients = append(clients, client)
	}
	for _, client := range clients {
		go client.RecvLockOrUnlock()
	}
	for _, client := range clients {
		locker := client.Prepare()
		fmt.Printf("prepare %d %v\n", client.Clock.Pid, locker)
		if locker != nil {
			client.SendLock(locker)
			printClientClock(clients)
			client.SendUnLock(locker)
			printClientClock(clients)
			fmt.Println("")
		}
	}
}

func printClientClock(clients []*Client) {
	for _, client := range clients {
		fmt.Printf("%d ", client.Clock.Time)
	}
	fmt.Print("\n")
}
