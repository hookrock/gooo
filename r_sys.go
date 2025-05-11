package gooo

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// getSystemMemory 返回系统总内存大小（单位：字节）
func GetSystemMemory() (uint64, error) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxMemory()
	case "darwin":
		return getDarwinMemory()
	case "windows":
		return getWindowsMemory()
	default:
		return 0, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// 获取 Linux 系统内存
func getLinuxMemory() (uint64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				return 0, fmt.Errorf("invalid format")
			}
			memKB, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				return 0, err
			}
			return memKB * 1024, nil // KB → B
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, fmt.Errorf("meminfo not found")
}

// 获取 macOS 系统内存
func getDarwinMemory() (uint64, error) {
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	mem, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return 0, err
	}
	return mem, nil
}

// 获取 Windows 系统内存
func getWindowsMemory() (uint64, error) {
	cmd := exec.Command("wmic", "computersystem", "get", "TotalPhysicalMemory")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(output), "\r\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("invalid output")
	}
	memStr := strings.TrimSpace(lines[1])
	if memStr == "" {
		return 0, fmt.Errorf("empty value")
	}
	mem, err := strconv.ParseUint(memStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return mem, nil
}
