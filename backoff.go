package ghttp

import "time"

type Backoff interface {
	Reset()
	Next() time.Duration
}

type ConstantBackoff struct {
	Interval time.Duration
}

func (b *ConstantBackoff) Reset() {
}

func (b *ConstantBackoff) Next() time.Duration {
	return b.Interval
}

func NewConstantBackoff(d time.Duration) *ConstantBackoff {
	return &ConstantBackoff{Interval: d}
}
