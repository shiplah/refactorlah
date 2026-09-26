//go:build windows

package git

import "os"

func processAlive(pid int) (bool, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, nil
	}
	_ = process.Release()
	return true, nil
}
