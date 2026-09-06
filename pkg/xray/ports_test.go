package xray

import (
	"testing"

	"github.com/cockroachdb/errors"
	xc "github.com/miladrahimi/p-node/pkg/xray/config"
	"github.com/miladrahimi/p-node/pkg/xray/config/component"
)

func withInbounds(ports map[string]int) *xc.Config {
	c := xc.New("warning")
	c.Inbounds = nil
	for tag, port := range ports {
		c.Inbounds = append(c.Inbounds, &component.Inbound{Tag: tag, Port: port})
	}
	return c
}

func TestValidatePorts(t *testing.T) {
	busy := func(int) bool { return false }
	free := func(int) bool { return true }

	tests := []struct {
		name     string
		current  *xc.Config
		next     *xc.Config
		portFree func(int) bool
		wantErr  bool
	}{
		{"free ports pass", nil, withInbounds(map[string]int{"a": 1000, "b": 1001}), free, false},
		{"busy port fails", nil, withInbounds(map[string]int{"a": 1000}), busy, true},
		{"busy port held by us passes", withInbounds(map[string]int{"old": 1000}), withInbounds(map[string]int{"a": 1000}), busy, false},
		{"duplicate ports fail", nil, withInbounds(map[string]int{"a": 1000, "b": 1000}), free, true},
		{"api port is not checked", nil, withInbounds(map[string]int{"api": 1000}), busy, false},
		{"held api port does not cover others", withInbounds(map[string]int{"api": 1000}), withInbounds(map[string]int{"a": 1000}), busy, true},
		{"zero port is skipped", nil, withInbounds(map[string]int{"a": 0, "b": 0}), busy, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePorts(tt.current, tt.next, tt.portFree)
			if (err != nil) != tt.wantErr {
				t.Fatalf("got err %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ErrPortConflict) {
				t.Fatalf("error %v is not ErrPortConflict", err)
			}
		})
	}
}

func TestValidatePortsRejectsTwoApiInbounds(t *testing.T) {
	next := xc.New("warning")
	next.Inbounds = append(next.Inbounds, &component.Inbound{Tag: "api", Port: 2})
	if err := validatePorts(nil, next, func(int) bool { return true }); !errors.Is(err, ErrPortConflict) {
		t.Fatalf("expected ErrPortConflict, got %v", err)
	}
}
