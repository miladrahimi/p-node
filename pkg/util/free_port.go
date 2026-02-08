package util

import "net"

// FreePort finds a free port.
func FreePort() (int, error) {
	address, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		return 0, err
	}
	if err = listener.Close(); err != nil {
		return 0, err
	}

	return listener.Addr().(*net.TCPAddr).Port, err
}
