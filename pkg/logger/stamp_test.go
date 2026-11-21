package logger

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

type buffer struct {
	bytes.Buffer
	closed bool
	fail   bool
}

func (b *buffer) Write(p []byte) (int, error) {
	if b.fail {
		return 0, errors.New("disk full")
	}
	return b.Buffer.Write(p)
}

func (b *buffer) Close() error { b.closed = true; return nil }

func newTestStamp() (*Stamp, *buffer) {
	b := &buffer{}
	s := NewStamp(b)
	s.now = func() time.Time { return time.Date(2026, 11, 21, 9, 5, 12, 345e6, time.UTC) }
	return s, b
}

func TestStampPrefixesLines(t *testing.T) {
	s, b := newTestStamp()
	// A line split over writes, two lines in one write, and a dated paqet line.
	for _, p := range []string{
		"ssh: connect to host 1.2.3.4 ",
		"port 22: Connection timed out\nfirst\nsecond\n",
		"2026/11/21 09:09:00 shutdown signal received, shutting down...\n",
		"partial",
	} {
		if n, err := s.Write([]byte(p)); n != len(p) || err != nil {
			t.Fatalf("Write(%q) = %d, %v", p, n, err)
		}
	}
	if err := s.Close(); err != nil || !b.closed {
		t.Fatalf("Close = %v, closed %v", err, b.closed)
	}

	want := "2026-11-21 09:05:12.345 ssh: connect to host 1.2.3.4 port 22: Connection timed out\n" +
		"2026-11-21 09:05:12.345 first\n" +
		"2026-11-21 09:05:12.345 second\n" +
		"2026/11/21 09:09:00 shutdown signal received, shutting down...\n" +
		"2026-11-21 09:05:12.345 partial\n"
	if b.String() != want {
		t.Fatalf("got:\n%s\nwant:\n%s", b.String(), want)
	}
}

func TestStampSplitsLongLines(t *testing.T) {
	s, b := newTestStamp()
	_, _ = s.Write([]byte(strings.Repeat("x", maxStampLine+10)))
	if got := strings.Count(b.String(), "\n"); got != 1 {
		t.Fatalf("a line over the cap should be written at once, got %d lines", got)
	}
	_ = s.Close()
	if got := strings.Count(b.String(), "2026-11-21 09:05:12.345 "); got != 1 {
		t.Fatalf("the whole line goes out at the cap, got %d stamps", got)
	}
}

func TestStampSwallowsWriteErrors(t *testing.T) {
	s, b := newTestStamp()
	b.fail = true
	if n, err := s.Write([]byte("lost\n")); n != 5 || err != nil {
		t.Fatalf("Write = %d, %v; a log error must not reach the process pipe", n, err)
	}
}
