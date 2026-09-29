//go:build linux

package singularity

import "syscall"

func useIPTransparent(network, address string, conn syscall.RawConn) error {
	return conn.Control(func(descriptor uintptr) {
		syscall.SetsockoptInt(int(descriptor), syscall.IPPROTO_IP, syscall.IP_TRANSPARENT, 1)
	})
}
