package logical_clock

import (
	"sort"
	"sync"
)

type Client struct {
	Peers   []Connection
	Local   Connection
	Clock   *Clock
	lockers []Clock
	mutex   sync.Mutex
}

func (c *Client) lock(locker *Clock) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.lockers = append(c.lockers, *locker)
	sort.SliceStable(c.lockers, func(i, j int) bool {
		if c.lockers[i].Time < c.lockers[j].Time {
			return true
		} else if c.lockers[i].Time == c.lockers[j].Time {
			return c.lockers[i].Pid < c.lockers[j].Pid
		}
		return false
	})
}

func (c *Client) unlock(locker *Clock) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for i, clock := range c.lockers {
		if *locker == clock {
			c.lockers = append(c.lockers[:i], c.lockers[i+1:]...)
			break
		}
	}
}

func (c *Client) RecvLockOrUnlock() {
	for {
		func() {
			var message, ack = c.Local.Recv()
			defer ack()
			switch message.Type {
			case MessageTypePrepare:
				return
			case MessageTypeLock:
				c.lock(message.Clock)
				c.Clock.update(nil)
			case MessageTypeUnLock:
				c.unlock(<-message.Carrier)
			}
			c.Clock.update(message.Clock)
		}()
	}
}

func (c *Client) SendLock(locker *Clock) {
	var maxAck *Clock
	c.Clock.update(nil)
	for i := range c.Peers {
		if c.Peers[i].Pid() == c.Clock.Pid {
			continue
		}
		var message = Message{
			Type:  MessageTypeLock,
			Clock: locker,
		}
		ack := c.Peers[i].Send(&message)
		if maxAck == nil || ack.Time > maxAck.Time {
			maxAck = &ack
		}
	}
	if maxAck != nil {
		c.Clock.update(maxAck)
	}
	c.lock(locker)
}

func (c *Client) SendUnLock(locker *Clock) {
	c.Clock.update(nil)
	var clock = *c.Clock
	for i := range c.Peers {
		if c.Peers[i].Pid() == c.Clock.Pid {
			continue
		}
		var message = Message{
			Type:    MessageTypeUnLock,
			Clock:   &clock,
			Carrier: make(chan *Clock, 1),
		}
		message.Carrier <- locker
		_ = c.Peers[i].Send(&message)
	}
	c.unlock(locker)
}

func (c *Client) Prepare() *Clock {
	var locker = *c.Clock
	var maxAck *Clock
	for i := range c.Peers {
		if c.Peers[i].Pid() == c.Clock.Pid {
			continue
		}
		var message = Message{
			Type: MessageTypePrepare,
		}
		ack := c.Peers[i].Send(&message)
		if maxAck == nil || ack.Time > maxAck.Time {
			maxAck = &ack
		}
	}
	if maxAck == nil || c.Clock.Time >= maxAck.Time {
		return &locker
	}
	return nil
}
