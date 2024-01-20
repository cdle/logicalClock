package logical_clock

type Clock struct {
	Pid  int
	Time int64
}

func (c *Clock) update(clock *Clock) {
	if clock == nil {
		c.Time++
	} else {
		c.Time = max(c.Time, clock.Time) + 1
	}
}
