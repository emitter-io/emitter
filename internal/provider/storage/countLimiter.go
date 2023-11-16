package storage

import "github.com/emitter-io/emitter/internal/message"

type Limiter interface {
	Admit(*message.Message) bool
	Limit(*message.Frame)
}

// MessageCountLimiter provide an Limiter implementation to replace the "limit"
// parameter in the Query() function.
type MessageCountLimiter struct {
	count    int64 `binary:"-"`
	MsgLimit int64 // TODO: why is this exported?
}

func (limiter *MessageCountLimiter) Admit(m *message.Message) bool {
	// As this function won't be called multiple times once the limit is reached,
	// the following implementation should be faster than using a branching statement
	// to check if the limit is reached, before incrementing the counter.
	limiter.count += 1
	return limiter.count <= limiter.MsgLimit

	// The following implementation would use a branching each time the function is called.
	/*
		if limiter.count < limiter.MsgLimit {
			limiter.count += 1
			return true
		}
		return false
	*/
}

func (limiter *MessageCountLimiter) Limit(frame *message.Frame) {
	frame.Limit(int(limiter.MsgLimit))
}

func NewMessageNumberLimiter(limit int64) Limiter {
	return &MessageCountLimiter{MsgLimit: limit}
}
