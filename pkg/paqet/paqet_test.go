package paqet

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/miladrahimi/p-node/pkg/logger"
)

// The log file gets a time on every line: paqet's own lines keep theirs, and
// anything else it prints (a panic) is stamped.
func TestPaqetLogLinesHaveTime(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "storage", "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	l, err := logger.New("debug", logger.StampLayout, make(chan struct{}))
	if err != nil {
		t.Fatal(err)
	}

	binary := filepath.Join(dir, "paqet")
	script := "#!/bin/sh\necho '2026/11/21 09:09:00 dated by paqet'\necho 'panic: undated' >&2\nexec sleep 30\n"
	if err := os.WriteFile(binary, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(dir, "paqet.log")
	p := New(l, binary, filepath.Join(dir, "paqet.yaml"), logPath)
	network := Network{Interface: "eth0", Address: "10.0.0.2", RouterMac: "aa:bb:cc:dd:ee:ff"}
	if err := p.Reconfigure(ClientConfig(network, 1080, "1.2.3.4", 4443, "key")); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		b, _ := os.ReadFile(logPath)
		if strings.Count(string(b), "\n") >= 2 || time.Now().After(deadline) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err := p.Stop(); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(b)), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %q", b)
	}
	stamped := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3} panic: undated$`)
	for _, line := range lines {
		switch {
		case line == "2026/11/21 09:09:00 dated by paqet":
		case stamped.MatchString(line):
		default:
			t.Errorf("line without the expected time: %q", line)
		}
	}
}
