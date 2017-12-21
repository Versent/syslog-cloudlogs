package batching

import (
	"time"

	"github.com/apex/log"
	"github.com/sirupsen/logrus"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
)

// DispatchFunc invoked when a batch is ready to send
type DispatchFunc func([]*LogEntry)

// LogEntry decoded log entry
type LogEntry struct {
	Message        string
	MilliTimestamp int64
}

// Batcher builds lists of records for dispatch
type Batcher struct {
	dispatchFunc DispatchFunc
	records      []*LogEntry
	flushTimer   *time.Timer
	size         int
	capacity     int
	duration     time.Duration
}

// NewBatcher configure a new batcher and it's dipsatch function
func NewBatcher(capacity int, duration time.Duration, dispatchFunc DispatchFunc) *Batcher {
	return &Batcher{
		dispatchFunc: dispatchFunc,
		records:      []*LogEntry{},
		capacity:     capacity,
		duration:     duration,
	}
}

// Handler handle incoming log messages and write batches to the dispatcher function
func (b *Batcher) Handler(channel syslog.LogPartsChannel) {
	logrus.Info("handle ready")

	b.flushTimer = time.NewTimer(b.duration)

	for {
		select {
		case logParts := <-channel:

			entry := &LogEntry{
				Message:        logParts["content"].(string),
				MilliTimestamp: makeMilliTimestamp(logParts["timestamp"].(time.Time)),
			}

			if b.willOverflow(len(entry.Message)) {
				log.Debugf("Batch flushed to prevent size overflow - size: %d, capacity: %v", b.size, b.capacity)
				b.flush()
			}

			b.records = append(b.records, entry)
			b.size += len(entry.Message)

			if b.isFullSize() {
				logrus.Debugf("Batch flushed due to batch size - size: %d, capacity: %v", b.size, b.capacity)
				b.flush()
			}

		case <-b.flushTimer.C:
			log.Debugf("Batch flushed due to timer - length: %v", len(b.records))
			b.flush()
		}
	}

}

// Length returns the current length of the buffer.
func (b *Batcher) Length() int {
	return len(b.records)
}

func (b *Batcher) willOverflow(size int) bool {
	return b.size+size > b.capacity
}

func (b *Batcher) isFullSize() bool {
	return b.size >= b.capacity
}

func (b *Batcher) flush() {
	records := b.records

	b.flushTimer = time.NewTimer(b.duration)
	b.records = []*LogEntry{}
	b.size = 0

	if len(records) != 0 {
		b.dispatchFunc(records)
	}
}

func makeMilliTimestamp(input time.Time) int64 {
	return input.UTC().UnixNano() / int64(time.Millisecond)
}
