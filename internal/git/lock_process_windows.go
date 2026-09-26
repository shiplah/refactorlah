//go:build windows

package git

import (
	"errors"
	"os"
	"syscall"
)

func processAlive(pid int) (bool, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		if errors.Is(err, syscall.ERROR_ACCESS_DENIED) {
			return true, nil
		}
		// OpenProcess reports an absent PID as ERROR_INVALID_PARAMETER (87).
		if errors.Is(err, syscall.Errno(87)) || errors.Is(err, syscall.ERROR_NOT_FOUND) {
			return false, nil
		}
		return false, err
	}
	_ = process.Release()
	return true, nil
}
