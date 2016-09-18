package sql

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/coreos/dex/storage"
	"github.com/coreos/dex/storage/conformance"
)

func withTimeout(t time.Duration, f func()) {
	c := make(chan struct{})
	defer close(c)

	go func() {
		select {
		case <-c:
		case <-time.After(t):
			// Dump a stack trace of the program. Useful for debugging deadlocks.
			buf := make([]byte, 2<<20)
			fmt.Fprintf(os.Stderr, "%s\n", buf[:runtime.Stack(buf, true)])
			panic("test took too long")
		}
	}()

	f()
}

func TestStorage(t *testing.T) {
	newStorage := func() storage.Storage {
		// NOTE(ericchiang): In memory means we only get one connection at a time. If we
		// ever write tests that require using multiple connections, for instance to test
		// transactions, we need to move to a file based system.
		s := &SQLite{":memory:"}
		conn, err := s.open()
		if err != nil {
			t.Fatal(err)
		}
		return conn
	}

	withTimeout(time.Second*10, func() {
		conformance.RunTestSuite(t, newStorage)
	})
}
