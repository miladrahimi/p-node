package paqet

import (
	"fmt"
	"net"
	"strconv"

	"github.com/cockroachdb/errors"
)

// Mtu is the KCP packet size. Iranian carriers fragment raw TCP/KCP packets
// above roughly 1280 bytes on the wire, which KCP handles badly, so stay
// comfortably below (1200 + 66 bytes of Ethernet/IP/TCP headers).
const Mtu = 1200

// Network is what paqet needs to craft raw packets on this host: the interface,
// its IPv4 address, and the gateway MAC the packets are sent to.
type Network struct {
	Interface string
	Address   string
	RouterMac string
}

// Validate checks the network.
func (n *Network) Validate() error {
	if n == nil {
		return errors.New("paqet: network is nil")
	}
	if n.Interface == "" {
		return errors.New("paqet: network interface is required")
	}
	if ip := net.ParseIP(n.Address); ip == nil || ip.To4() == nil {
		return errors.Errorf("paqet: network address %q is not an IPv4 address", n.Address)
	}
	if _, err := net.ParseMAC(n.RouterMac); err != nil {
		return errors.Errorf("paqet: router mac %q is invalid", n.RouterMac)
	}
	return nil
}

// Config is the paqet configuration for one process: a server (listens on Port)
// or a client (dials ServerAddr and exposes a SOCKS5 proxy on SocksPort). Both
// sides share Key. It is comparable, so a changed config is detected with ==.
type Config struct {
	Network    Network
	Key        string
	Port       int    // server: raw packet listen port
	ServerAddr string // client: paqet server host:port
	SocksPort  int    // client: local SOCKS5 port (127.0.0.1)
}

// ServerConfig makes a server config.
func ServerConfig(network Network, port int, key string) *Config {
	return &Config{Network: network, Key: key, Port: port}
}

// ClientConfig makes a client config.
func ClientConfig(network Network, socksPort int, serverHost string, serverPort int, key string) *Config {
	return &Config{
		Network:    network,
		Key:        key,
		ServerAddr: net.JoinHostPort(serverHost, strconv.Itoa(serverPort)),
		SocksPort:  socksPort,
	}
}

// Validate checks the config.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("paqet: config is nil")
	}
	if err := c.Network.Validate(); err != nil {
		return err
	}
	if c.Key == "" {
		return errors.New("paqet: key is required")
	}
	if c.Port > 0 {
		if c.ServerAddr != "" || c.SocksPort != 0 {
			return errors.New("paqet: a server config cannot have client fields")
		}
		return nil
	}
	if c.ServerAddr == "" {
		return errors.New("paqet: server address is required")
	}
	if c.SocksPort < 1 || c.SocksPort > 65535 {
		return errors.New("paqet: socks port is invalid")
	}
	return nil
}

// Yaml renders the paqet YAML config file.
func (c *Config) Yaml() string {
	n := c.Network
	if c.Port > 0 {
		return fmt.Sprintf(`role: "server"
log:
  level: "info"
listen:
  addr: ":%d"
network:
  interface: %q
  ipv4:
    addr: %q
    router_mac: %q
transport:
  protocol: "kcp"
  kcp:
    mtu: %d
    block: "aes"
    key: %q
`, c.Port, n.Interface, net.JoinHostPort(n.Address, strconv.Itoa(c.Port)), n.RouterMac, Mtu, c.Key)
	}
	return fmt.Sprintf(`role: "client"
log:
  level: "info"
socks5:
  - listen: %q
network:
  interface: %q
  ipv4:
    addr: %q
    router_mac: %q
server:
  addr: %q
transport:
  protocol: "kcp"
  kcp:
    mtu: %d
    block: "aes"
    key: %q
`, net.JoinHostPort("127.0.0.1", strconv.Itoa(c.SocksPort)), n.Interface, net.JoinHostPort(n.Address, "0"), n.RouterMac, c.ServerAddr, Mtu, c.Key)
}
