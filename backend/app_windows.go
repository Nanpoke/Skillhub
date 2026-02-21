//go:build windows

package backend

import (
	"syscall"
	"unsafe"
)

// getDiskFreeSpaceWindows Windows 获取磁盘空间（总空间 + 可用空间）
func getDiskFreeSpaceWindows(path string) (totalBytes, freeBytes uint64, err error) {
	kernel32, err := syscall.LoadLibrary("Kernel32.dll")
	if err != nil {
		return 0, 0, err
	}
	defer syscall.FreeLibrary(kernel32)

	getDiskFreeSpaceEx, err := syscall.GetProcAddress(kernel32, "GetDiskFreeSpaceExW")
	if err != nil {
		return 0, 0, err
	}

	// 将路径转换为 UTF-16
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0, err
	}

	var totalFreeBytes uint64

	// 调用 Windows API
	_, _, err = syscall.Syscall6(uintptr(getDiskFreeSpaceEx), 4,
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytes)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
		0, 0)

	if err != syscall.Errno(0) {
		return 0, 0, err
	}

	return totalBytes, freeBytes, nil
}

// getDiskFreeSpaceUnix Unix/Linux/macOS 获取磁盘空间
// Windows 版本的空实现（不会被调用）
func getDiskFreeSpaceUnix(path string) (totalBytes, freeBytes uint64, err error) {
	return 0, 0, syscall.EWINDOWS
}
