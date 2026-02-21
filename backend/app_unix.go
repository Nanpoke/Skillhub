//go:build darwin || linux || freebsd || netbsd || openbsd

package backend

import (
	"syscall"
)

// getDiskFreeSpaceUnix Unix/Linux/macOS 获取磁盘空间（总空间 + 可用空间）
func getDiskFreeSpaceUnix(path string) (totalBytes, freeBytes uint64, err error) {
	var stat syscall.Statfs_t
	err = syscall.Statfs(path, &stat)
	if err != nil {
		return 0, 0, err
	}
	// 总空间 = Blocks * Bsize
	// 可用空间 = Bavail * Bsize
	totalBytes = stat.Blocks * uint64(stat.Bsize)
	freeBytes = stat.Bavail * uint64(stat.Bsize)
	return totalBytes, freeBytes, nil
}

// getDiskFreeSpaceWindows Windows 获取磁盘空间
// Unix 版本的空实现（不会被调用）
func getDiskFreeSpaceWindows(path string) (totalBytes, freeBytes uint64, err error) {
	return 0, 0, syscall.EOPNOTSUPP
}
