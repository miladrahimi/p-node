package util

import (
	"fmt"
	"net"
)

// PortFree checks if the given port is free or not.
func PortFree(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}

	if err = listener.Close(); err != nil {
		return false
	}

	return true
}
