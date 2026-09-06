package util

import (
	"bytes"
	"strings"

	"github.com/cockroachdb/errors"
)

// ParseX25519Output parses output of `xray x25519` and returns private key and password (public key).
func ParseX25519Output(output []byte) (string, string, error) {
	privateKey := ""
	password := ""
	lines := strings.Split(string(bytes.TrimSpace(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		// xray labels the lines "PrivateKey" and "Password (PublicKey)".
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		switch {
		case strings.HasPrefix(key, "privatekey"):
			privateKey = value
		case strings.HasPrefix(key, "password"):
			password = value
		}
	}

	if privateKey == "" || password == "" {
		return "", "", errors.New("xray: invalid x25519 output")
	}
	return privateKey, password, nil
}
