package cwlogs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/versent/syslog-cloudlogs/pkg/batching"
	"github.com/wolfeidau/go-syslog/format"
)

func TestTransformEntriesToEvents(t *testing.T) {

	dispatcher := &Dispatcher{}

	le := []*batching.LogEntry{
		&batching.LogEntry{
			Parts: format.LogParts{
				"content":   "Mon Mar 05 05:23:14 UTC 2018Info: { \"Time\": \"Mon, 5 Mar 2018 05:23:14 UTC\" }\u0000",
				"timestamp": time.Now(),
			},
		},
	}

	events := dispatcher.transformEntriesToEvents(le)

	require.Len(t, events, 1)
}
