package paqet

import (
	"os/exec"
	"runtime"
	"strconv"

	"github.com/miladrahimi/p-node/pkg/logger"
	"go.uber.org/zap"
)

// A paqet server reads packets with pcap, below the kernel TCP stack, so the
// kernel must neither track those flows nor answer them with RST. These are
// the rules the paqet README requires; they are added for a server port when
// it starts and removed when it stops. Linux only, best effort.
func firewallRules(port int) [][]string {
	p := strconv.Itoa(port)
	return [][]string{
		{"-t", "raw", "PREROUTING", "-p", "tcp", "--dport", p, "-j", "NOTRACK"},
		{"-t", "raw", "OUTPUT", "-p", "tcp", "--sport", p, "-j", "NOTRACK"},
		{"-t", "mangle", "OUTPUT", "-p", "tcp", "--sport", p, "--tcp-flags", "RST", "RST", "-j", "DROP"},
	}
}

func allowPort(l *logger.Logger, port int) {
	iptables, ok := findIptables(l)
	if !ok {
		return
	}
	for _, rule := range firewallRules(port) {
		if runIptables(iptables, rule, "-C") == nil {
			continue
		}
		if err := runIptables(iptables, rule, "-A"); err != nil {
			l.Error("paqet: cannot add firewall rule", zap.Strings("rule", rule), zap.Error(err))
		}
	}
}

func disallowPort(l *logger.Logger, port int) {
	iptables, ok := findIptables(l)
	if !ok {
		return
	}
	for _, rule := range firewallRules(port) {
		for runIptables(iptables, rule, "-C") == nil {
			if err := runIptables(iptables, rule, "-D"); err != nil {
				l.Error("paqet: cannot remove firewall rule", zap.Strings("rule", rule), zap.Error(err))
				break
			}
		}
	}
}

func findIptables(l *logger.Logger) (string, bool) {
	if runtime.GOOS != "linux" {
		return "", false
	}
	path, err := exec.LookPath("iptables")
	if err != nil {
		l.Error("paqet: iptables not found; the kernel may reset paqet server connections")
		return "", false
	}
	return path, true
}

// runIptables runs `iptables -t <table> <op> <chain> <match...>` where op is
// -C (check), -A (append), or -D (delete).
func runIptables(iptables string, rule []string, op string) error {
	args := append([]string{}, rule[:2]...)
	args = append(args, op)
	args = append(args, rule[2:]...)
	return exec.Command(iptables, args...).Run()
}
