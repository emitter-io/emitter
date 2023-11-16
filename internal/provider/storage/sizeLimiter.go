package storage

import "github.com/emitter-io/emitter/internal/message"

// MessageSizeLimiter provide an Limiter implementation based on both the
// number of messages and the total size of the response.
type MessageSizeLimiter struct {
	count      int64 `binary:"-"`
	size       int64 `binary:"-"`
	countLimit int64
	sizeLimit  int64
}

func (limiter *MessageSizeLimiter) Admit(m *message.Message) bool {
	// As this function won't be called multiple times once the limit is reached,
	// the following implementation should be faster than using a branching statement
	// to check if the limit is reached, before incrementing the counter.
	// Todo: discuss whether it's ok to make that assumption

	// This size calculation comes from mqtt.go:EncodeTo() line 320.
	// Todo: discuss whether this is the best way to calculate the size.
	// 2 bytes for message ID.
	limiter.size += int64(2 + len(m.Channel) + len(m.Payload))
	limiter.count += 1
	return limiter.count <= limiter.countLimit && limiter.size <= limiter.sizeLimit
}

func (limiter *MessageSizeLimiter) Limit(frame *message.Frame) {
	// Limit takes the first N elements that fit into a message, sorted by message time
	frame.Sort()
	frame.Limit(int(limiter.countLimit))

	totalSize := int64(0)
	for i, m := range *frame {
		totalSize += int64(2 + len(m.Channel) + len(m.Payload))
		if totalSize > limiter.sizeLimit {
			*frame = (*frame)[:i]
			break
		}
		limiter.size += totalSize
	}
}

func NewMessageSizeLimiter(countLimit, sizeLimit int64) Limiter {
	return &MessageSizeLimiter{countLimit: countLimit, sizeLimit: sizeLimit}
}
