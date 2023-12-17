package types

import "time"

// RetryCounter is a helper struct that ensure retrying for at least N times and not expire before time T
type RetryCounter struct {
	deadlineT  int64
	minTryN    int8
	totalTried int8
}

func NewRetryCounter(expireAfterSecs int64, minRetry int8) *RetryCounter {
	if minRetry < 1 {
		panic("invalid min retry")
	}
	return &RetryCounter{
		deadlineT:  time.Now().UTC().Unix() + expireAfterSecs,
		minTryN:    minRetry + 1,
		totalTried: 0,
	}
}

// DefaultRetryCounterFetchingRpc returns default counter, ensure at least retry for 5 times, minimum 10s
func DefaultRetryCounterFetchingRpc() *RetryCounter {
	return NewRetryCounter(10, 5)
}

func (c *RetryCounter) Continue() bool {
	c.totalTried++
	if c.totalTried < c.minTryN {
		return true
	}
	return c.deadlineT > time.Now().UTC().Unix()
}
