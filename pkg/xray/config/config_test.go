package config

import (
	"testing"

	"github.com/miladrahimi/p-node/pkg/xray/config/component"
)

func TestEqualsIgnoresMetadata(t *testing.T) {
	a := New("warning")
	b := New("warning")
	a.Metadata = &component.Metadata{UpdatedAt: "2026-10-27T15:00:00Z", UpdatedBy: "m1"}
	b.Metadata = &component.Metadata{UpdatedAt: "2026-10-27T16:00:00Z", UpdatedBy: "m2"}
	if !a.Equals(b) {
		t.Error("configs differing only in _metadata must be equal")
	}
	if !a.Equals(New("warning")) {
		t.Error("a config with _metadata must equal the same config without it")
	}
	if a.Metadata == nil || b.Metadata == nil {
		t.Error("Equals must not clear the metadata of its operands")
	}

	b.Log.LogLevel = "info"
	if a.Equals(b) {
		t.Error("configs differing outside _metadata must not be equal")
	}

	var nilConfig *Config
	if a.Equals(nilConfig) || nilConfig.Equals(a) {
		t.Error("a config never equals nil")
	}
	if !nilConfig.Equals(nil) {
		t.Error("nil equals nil")
	}
}
