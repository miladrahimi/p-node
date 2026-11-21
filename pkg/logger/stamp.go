package logger

import (
	"bytes"
	"io"
	"regexp"
	"sync"
	"time"
)

// StampLayout is the time layout of stamped lines, the app logs' default format.
const StampLayout = "2006-01-02 15:04:05.000"

// maxStampLine caps a buffered line without a newline; a longer one is written in parts.
const maxStampLine = 64 << 10

// datedLine matches a line that already starts with a date and time, as Go's
// log package (paqet) and Xray write them.
var datedLine = regexp.MustCompile(`^\d{4}[-/]\d{2}[-/]\d{2}[ T]\d{2}:\d{2}:\d{2}`)

// Stamp writes a child process's output (ssh, paqet) to its log file with the
// time at the start of every line; lines that already start with one are kept
// as they are. Write never fails: a log write error must not break the pipe
// and with it the process.
type Stamp struct {
	mu   sync.Mutex
	w    io.WriteCloser
	now  func() time.Time
	line []byte
}

// NewStamp wraps w; Close writes a trailing partial line and closes w.
func NewStamp(w io.WriteCloser) *Stamp {
	return &Stamp{w: w, now: time.Now}
}

// Write buffers p and writes every complete line, stamped.
func (s *Stamp) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.line = append(s.line, p...)
	start := 0
	for {
		i := bytes.IndexByte(s.line[start:], '\n')
		if i < 0 {
			break
		}
		s.writeLine(s.line[start : start+i+1])
		start += i + 1
	}
	s.line = append(s.line[:0], s.line[start:]...)
	if len(s.line) >= maxStampLine {
		s.writeLine(append(s.line, '\n'))
		s.line = s.line[:0]
	}
	return len(p), nil
}

// Close writes a trailing partial line and closes the underlying writer.
func (s *Stamp) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.line) > 0 {
		s.writeLine(append(s.line, '\n'))
		s.line = nil
	}
	return s.w.Close()
}

func (s *Stamp) writeLine(line []byte) {
	if !datedLine.Match(line) {
		line = append([]byte(s.now().Format(StampLayout)+" "), line...)
	}
	_, _ = s.w.Write(line)
}
