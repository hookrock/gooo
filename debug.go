package gooo

import "log"

// DebugMode 全局变量，标记是否启用调试模式
var DebugMode bool

// Debug 开启调试模式
func Debug() {
	DebugMode = true
}

// IsDebugMode 是否启用调试模式
func IsDebugMode() bool {
	return DebugMode
}

// debug.go
func DebugPrint(format string, v ...any) {
	if IsDebugMode() {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func SetDebugMode(b bool) {
	DebugMode = b
}
