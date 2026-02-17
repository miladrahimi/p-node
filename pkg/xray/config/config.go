package config

import (
	"encoding/json"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/miladrahimi/p-node/pkg/xray/config/component"
)

// Config represents the xray config struct.
type Config struct {
	Log       *component.Log        `json:"log" validate:"required"`
	Inbounds  []*component.Inbound  `json:"inbounds" validate:"required,dive"`
	Outbounds []*component.Outbound `json:"outbounds" validate:"required,dive"`
	Dns       *component.Dns        `json:"dns,omitempty" validate:"omitempty"`
	Stats     *component.Stats      `json:"stats,omitempty" validate:"omitempty"`
	API       *component.Api        `json:"api,omitempty" validate:"omitempty"`
	Policy    *component.Policy     `json:"policy,omitempty" validate:"omitempty"`
	Routing   *component.Routing    `json:"routing,omitempty" validate:"omitempty"`
	Reverse   *component.Reverse    `json:"reverse,omitempty" validate:"omitempty"`
	Metadata  *component.Metadata   `json:"_metadata,omitempty" validate:"omitempty"`
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
				Protocol: "tunnel",
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
		Dns: &component.Dns{
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
	if c.Routing == nil {
		return nil
	}
	for _, balancer := range c.Routing.Balancers {
		if balancer.Tag == tag {
			return balancer
		}
	}
	return nil
}

// Validate validates the xray config struct.
func (c *Config) Validate() error {
	if c.API != nil && c.FindInbound("api") == nil {
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
