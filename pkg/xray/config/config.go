package config

import (
	"encoding/json"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/miladrahimi/p-node/pkg/xray/config/component"
	"github.com/miladrahimi/p-node/pkg/xray/config/protocol"
)

// Config represents the xray config struct.
type Config struct {
	Log       *component.Log        `json:"log" validate:"required"`
	Inbounds  []*component.Inbound  `json:"inbounds" validate:"required,dive"`
	Outbounds []*component.Outbound `json:"outbounds" validate:"required,dive"`
	DNS       *component.Dns        `json:"dns" validate:"required"`
	Stats     *component.Stats      `json:"stats" validate:"required"`
	API       *component.Api        `json:"api" validate:"required"`
	Policy    *component.Policy     `json:"policy" validate:"required"`
	Routing   *component.Routing    `json:"routing" validate:"required"`
	Reverse   *component.Reverse    `json:"reverse,omitempty"`
	Metadata  *component.Metadata   `json:"_metadata,omitempty"`
}

// New creates a new xray config struct.
func New(logLevel string) *Config {
	return &Config{
		Log: &component.Log{
			LogLevel: logLevel,
			Access:   "./storage/logs/xray-access.log",
			Error:    "./storage/logs/xray-error.log",
		},
		Inbounds: []*component.Inbound{
			{
				Tag:      "api",
				Protocol: "dokodemo-door",
				Listen:   "127.0.0.1",
				Port:     3411,
				Settings: &component.InboundSettings{
					Address: "127.0.0.1",
					Network: "tcp",
				},
			},
		},
		Outbounds: []*component.Outbound{
			{
				Tag:      "out",
				Protocol: "freedom",
			},
		},
		DNS: &component.Dns{
			Servers: []string{"8.8.8.8", "8.8.4.4", "localhost"},
		},
		Stats: &component.Stats{},
		API: &component.Api{
			Tag:      "api",
			Services: []string{"StatsService"},
		},
		Policy: &component.Policy{
			Levels: map[string]component.PolicyLevel{
				"0": {
					StatsUserUplink:   true,
					StatsUserDownlink: true,
				},
			},
			System: component.PolicySystem{
				StatsInboundUplink:    true,
				StatsInboundDownlink:  true,
				StatsOutboundUplink:   true,
				StatsOutboundDownlink: true,
			},
		},
		Routing: &component.Routing{
			DomainStrategy: "AsIs",
			DomainMatcher:  "hybrid",
			Rules: []*component.Rule{
				{
					InboundTag:  []string{"api"},
					OutboundTag: "api",
				},
			},
			Balancers: []*component.Balancer{},
		},
		Reverse: &component.Reverse{
			Bridges: []*component.ReverseItem{},
			Portals: []*component.ReverseItem{},
		},
	}
}

// MakeShadowsocksInbound creates a shadowsocks inbound.
func (c *Config) MakeShadowsocksInbound(
	tag, password, method, network string,
	port int,
	ssClients []*protocol.SsClient,
) *component.Inbound {
	clients := make([]any, 0, len(ssClients))
	for _, client := range ssClients {
		clients = append(clients, client)
	}

	return &component.Inbound{
		Tag:      tag,
		Protocol: "shadowsocks",
		Listen:   "0.0.0.0",
		Port:     port,
		Settings: &component.InboundSettings{
			Clients:  clients,
			Password: password,
			Method:   method,
			Network:  network,
		},
	}
}

// MakeShadowsocksOutbound creates a shadowsocks outbound.
func (c *Config) MakeShadowsocksOutbound(tag, host, password, method string, port int) *component.Outbound {
	return &component.Outbound{
		Tag:      tag,
		Protocol: "shadowsocks",
		Settings: &protocol.SsOutboundSettings{
			Servers: []*protocol.SsOutboundServer{
				{
					Address:  host,
					Port:     port,
					Method:   method,
					Password: password,
					Uot:      true,
				},
			},
		},
		StreamSettings: &component.StreamSettings{
			Network: "tcp",
		},
	}
}

// FindInbound finds an inbound by tag.
func (c *Config) FindInbound(tag string) *component.Inbound {
	for _, inbound := range c.Inbounds {
		if inbound.Tag == tag {
			return inbound
		}
	}
	return nil
}

// FindOutbound finds an outbound by tag.
func (c *Config) FindOutbound(tag string) *component.Outbound {
	for _, outbound := range c.Outbounds {
		if outbound.Tag == tag {
			return outbound
		}
	}
	return nil
}

// FindBalancer finds a balancer by tag.
func (c *Config) FindBalancer(tag string) *component.Balancer {
	for _, balancer := range c.Routing.Balancers {
		if balancer.Tag == tag {
			return balancer
		}
	}
	return nil
}

// Validate validates the xray config struct.
func (c *Config) Validate() error {
	if c.FindInbound("api") == nil {
		return errors.New("xray: config: api inbound not found")
	}
	return errors.WithStack(validator.New(validator.WithRequiredStructEnabled()).Struct(c))
}

// Equals checks if two config structs are equal.
func (c *Config) Equals(other *Config) bool {
	json1, err := json.Marshal(c)
	if err != nil {
		return false
	}

	json2, err := json.Marshal(other)
	if err != nil {
		return false
	}

	return string(json1) == string(json2)
}
