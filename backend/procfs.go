package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func getProcessCwd(pid int) (string, error) {
	switch runtime.GOOS {
	case "linux":
		return os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid))
	case "darwin":
		out, err := exec.Command("lsof", "-a", "-p", strconv.Itoa(pid), "-d", "cwd", "-Fn").Output()
		if err != nil {
			return "", fmt.Errorf("lsof: %w", err)
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, "n") {
				return line[1:], nil
			}
		}
		return "", fmt.Errorf("cwd not found for pid %d", pid)
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}
