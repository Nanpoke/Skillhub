package utils

import (
	"fmt"
	"path/filepath"
	"strings"
)

// SanitizeZipPath 验证 ZIP 解压路径安全性
// 防止路径遍历攻击（Zip Slip 漏洞）
func SanitizeZipPath(destDir, fileName string) (string, error) {
	// 检查路径遍历
	if strings.Contains(fileName, "..") {
		return "", fmt.Errorf("invalid path: contains '..'")
	}

	// 检查绝对路径
	if filepath.IsAbs(fileName) {
		return "", fmt.Errorf("invalid path: absolute path not allowed")
	}

	// 检查 Windows 盘符路径 (如 C:\)
	if len(fileName) >= 2 && fileName[1] == ':' {
		return "", fmt.Errorf("invalid path: Windows absolute path not allowed")
	}

	// 构建完整路径
	fpath := filepath.Join(destDir, fileName)

	// 获取绝对路径进行最终验证
	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute dest path: %w", err)
	}

	absPath, err := filepath.Abs(fpath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute file path: %w", err)
	}

	// 验证最终路径仍在目标目录内
	if !strings.HasPrefix(absPath, absDest+string(filepath.Separator)) && absPath != absDest {
		return "", fmt.Errorf("path traversal detected: %s", fileName)
	}

	return fpath, nil
}

// ValidateSkillName 验证 Skill 名称安全性
// 防止路径遍历攻击
func ValidateSkillName(name string) error {
	if name == "" {
		return fmt.Errorf("skill name cannot be empty")
	}

	// 检查路径遍历
	if strings.Contains(name, "..") {
		return fmt.Errorf("invalid skill name: contains '..'")
	}

	// 检查路径分隔符
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("invalid skill name: contains path separator")
	}

	// 检查绝对路径
	if filepath.IsAbs(name) {
		return fmt.Errorf("invalid skill name: absolute path not allowed")
	}

	// 检查 Windows 盘符路径
	if len(name) >= 2 && name[1] == ':' {
		return fmt.Errorf("invalid skill name: Windows absolute path not allowed")
	}

	// 检查 null 字节
	if strings.ContainsRune(name, '\x00') {
		return fmt.Errorf("invalid skill name: contains null byte")
	}

	return nil
}

// ValidatePathInDir 验证路径是否在指定目录内
func ValidatePathInDir(baseDir, targetPath string) error {
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute base path: %w", err)
	}

	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute target path: %w", err)
	}

	if !strings.HasPrefix(absTarget, absBase+string(filepath.Separator)) && absTarget != absBase {
		return fmt.Errorf("path escapes base directory")
	}

	return nil
}
