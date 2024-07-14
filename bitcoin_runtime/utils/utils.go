package utils

import (
	"os/exec"
	"strings"
)

func IsProcessRunning(name string) bool {
	cmd := exec.Command("pgrep", name)
	out, err := cmd.Output()

	if err != nil {
		return false
	}

	if len(strings.TrimSpace(string(out))) == 0 {
		return false
	}

	return true
}
