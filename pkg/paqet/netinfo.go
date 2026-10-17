package paqet

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
)

// DetectNetwork finds the default-route interface, its IPv4 address, and the
// gateway's MAC address, i.e. what paqet needs to send raw packets to the
// internet from this host.
func DetectNetwork() (*Network, error) {
	var iface, gateway string
	var err error
	switch runtime.GOOS {
	case "linux":
		iface, gateway, err = defaultRouteLinux()
	case "darwin":
		iface, gateway, err = defaultRouteDarwin()
	default:
		return nil, errors.Errorf("paqet: network detection is not supported on %s", runtime.GOOS)
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}

	address, err := interfaceIPv4(iface)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	mac, err := gatewayMac(iface, gateway)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Network{Interface: iface, Address: address, RouterMac: mac}, nil
}

// defaultRouteLinux reads the default route (interface and gateway) from /proc.
func defaultRouteLinux() (string, string, error) {
	file, err := os.Open("/proc/net/route")
	if err != nil {
		return "", "", errors.WithStack(err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		// Iface Destination Gateway Flags ...; destination 00000000 is the default route.
		if len(fields) < 3 || fields[1] != "00000000" {
			continue
		}
		raw, err := hex.DecodeString(fields[2])
		if err != nil || len(raw) != 4 {
			continue
		}
		// Little-endian hex.
		gateway := net.IPv4(raw[3], raw[2], raw[1], raw[0]).String()
		return fields[0], gateway, nil
	}
	return "", "", errors.New("paqet: no default route found")
}

// defaultRouteDarwin asks `route` for the default route.
func defaultRouteDarwin() (string, string, error) {
	out, err := exec.Command("route", "-n", "get", "default").CombinedOutput()
	if err != nil {
		return "", "", errors.Wrapf(err, "paqet: route failed: %s", strings.TrimSpace(string(out)))
	}
	var iface, gateway string
	for _, line := range strings.Split(string(out), "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		switch key {
		case "interface":
			iface = strings.TrimSpace(value)
		case "gateway":
			gateway = strings.TrimSpace(value)
		}
	}
	if iface == "" || gateway == "" {
		return "", "", errors.New("paqet: no default route found")
	}
	return iface, gateway, nil
}

// interfaceIPv4 returns the first IPv4 address of the interface.
func interfaceIPv4(name string) (string, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return "", errors.WithStack(err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return "", errors.WithStack(err)
	}
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}
	return "", errors.Errorf("paqet: interface %s has no IPv4 address", name)
}

// gatewayMac resolves the gateway's MAC from the neighbor table, poking the
// gateway first so an entry exists.
func gatewayMac(iface, gateway string) (string, error) {
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			time.Sleep(300 * time.Millisecond)
		}
		pokeGateway(gateway)

		var mac string
		var err error
		if runtime.GOOS == "linux" {
			mac, err = arpLinux(iface, gateway)
		} else {
			mac, err = arpDarwin(gateway)
		}
		if err == nil {
			return mac, nil
		}
		lastErr = err
	}
	return "", lastErr
}

// pokeGateway sends a UDP datagram towards the gateway to trigger ARP resolution.
func pokeGateway(gateway string) {
	conn, err := net.DialTimeout("udp", net.JoinHostPort(gateway, "9"), time.Second)
	if err != nil {
		return
	}
	_, _ = conn.Write([]byte{0})
	_ = conn.Close()
}

func arpLinux(iface, gateway string) (string, error) {
	file, err := os.Open("/proc/net/arp")
	if err != nil {
		return "", errors.WithStack(err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// IP address HW type Flags HW address Mask Device
		fields := strings.Fields(scanner.Text())
		if len(fields) < 6 || fields[0] != gateway || fields[5] != iface {
			continue
		}
		return normalizeMac(fields[3])
	}
	return "", errors.Errorf("paqet: gateway %s not found in the arp table", gateway)
}

var arpDarwinPattern = regexp.MustCompile(`\(([^)]+)\) at ([0-9a-fA-F:]+)`)

func arpDarwin(gateway string) (string, error) {
	out, err := exec.Command("arp", "-n", gateway).CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "paqet: arp failed: %s", strings.TrimSpace(string(out)))
	}
	if m := arpDarwinPattern.FindStringSubmatch(string(out)); m != nil && m[1] == gateway {
		return normalizeMac(m[2])
	}
	return "", errors.Errorf("paqet: gateway %s not found in the arp table", gateway)
}

// normalizeMac pads each octet (macOS prints "a:b:c:1:2:3") and rejects
// incomplete entries.
func normalizeMac(mac string) (string, error) {
	parts := strings.Split(mac, ":")
	if len(parts) != 6 {
		return "", errors.Errorf("paqet: invalid mac %q", mac)
	}
	for i, part := range parts {
		if len(part) == 1 {
			parts[i] = "0" + part
		}
	}
	normalized := strings.ToLower(strings.Join(parts, ":"))
	if normalized == "00:00:00:00:00:00" {
		return "", errors.Errorf("paqet: mac %q is incomplete", mac)
	}
	if _, err := net.ParseMAC(normalized); err != nil {
		return "", errors.Errorf("paqet: invalid mac %q", mac)
	}
	return normalized, nil
}

// String describes the network for logs.
func (n *Network) String() string {
	return fmt.Sprintf("%s %s via %s", n.Interface, n.Address, n.RouterMac)
}
