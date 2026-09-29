//go:build !linux

package singularity

import (
	"errors"
	"syscall"
)

func useIPTransparent(network, address string, conn syscall.RawConn) error {
	return errors.New("transparent proxy support requires Linux")
}
