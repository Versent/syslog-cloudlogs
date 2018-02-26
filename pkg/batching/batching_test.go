package batching

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	syslog "github.com/wolfeidau/go-syslog"
	"github.com/wolfeidau/go-syslog/format"
)

func Test_WhenNotFull(t *testing.T) {
	channel := make(syslog.LogPartsChannel)
	recordsChan := make(chan []*LogEntry, 1)
	batcher := NewBatcher(100, 1*time.Second, dispatch(recordsChan))

	go batcher.Handler(channel)

	channel <- format.LogParts{
		"content":   "test123",
		"timestamp": time.Now(),
	}

	require.Equal(t, 1, batcher.Length())
}

func Test_WhenFull(t *testing.T) {
	channel := make(syslog.LogPartsChannel)
	recordsChan := make(chan []*LogEntry, 1)
	batcher := NewBatcher(1, 1*time.Second, dispatch(recordsChan))

	go batcher.Handler(channel)

	channel <- format.LogParts{
		"content":   "test123",
		"timestamp": time.Now(),
	}

	records := <-recordsChan

	require.Equal(t, 0, batcher.Length())
	require.Len(t, records, 1)
}

func Test_WhenOverflow(t *testing.T) {
	channel := make(syslog.LogPartsChannel)
	recordsChan := make(chan []*LogEntry, 1)
	batcher := NewBatcher(10, 1*time.Second, dispatch(recordsChan))

	go batcher.Handler(channel)

	channel <- format.LogParts{
		"content":   "test123",
		"timestamp": time.Now(),
	}

	channel <- format.LogParts{
		"content":   "test12333",
		"timestamp": time.Now(),
	}

	records := <-recordsChan

	require.Equal(t, 1, batcher.Length())
	require.Len(t, records, 1)
}

func Test_WhenTimeout(t *testing.T) {
	channel := make(syslog.LogPartsChannel)
	recordsChan := make(chan []*LogEntry, 1)
	batcher := NewBatcher(100, 250*time.Millisecond, dispatch(recordsChan))

	go batcher.Handler(channel)

	channel <- format.LogParts{
		"content":   "test123",
		"timestamp": time.Now(),
	}

	records := <-recordsChan

	require.Equal(t, 0, batcher.Length())
	require.Len(t, records, 1)
}

func dispatch(records chan []*LogEntry) func([]*LogEntry) {
	return func(entries []*LogEntry) {
		records <- entries
	}
}
