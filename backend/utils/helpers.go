package utils

import (
	"os"
	stdruntime "runtime"
	"path/filepath"
	"strings"
)

// expandPath 展开路径中的 ~ 和环境变量（跨平台）
func ExpandPath(path string) string {
	// 处理空字符串边界情况
	if path == "" {
		return ""
	}

	// 处理 ~ 前缀
	if strings.HasPrefix(path, "~/") || path == "~" {
		home := GetHomeDir()
		if path == "~" {
			return home
		}
		return filepath.Join(home, path[2:])
	}

	// 处理 Windows 风格的路径
	if stdruntime.GOOS == "windows" {
		// 展开环境变量如 %USERPROFILE%
		path = os.ExpandEnv(path)
	}

	return path
}

// GetHomeDir 跨平台获取用户主目录
func GetHomeDir() string {
	if stdruntime.GOOS == "windows" {
		home := os.Getenv("USERPROFILE")
		if home == "" {
			home = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		}
		return home
	}
	return os.Getenv("HOME")
}
