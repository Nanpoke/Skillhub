package utils

import (
	"fmt"
	"os"
)

// CreateSkillLink 创建目录链接（Windows=Junction，Unix=Symlink）
// linkPath 是链接路径，targetPath 是源目录路径
func CreateSkillLink(linkPath, targetPath string) error {
	return createSkillLink(linkPath, targetPath)
}

// IsSkillLink 检测路径是否为链接（Junction 或 Symlink）
func IsSkillLink(path string) (bool, error) {
	return isSkillLink(path)
}

// RemoveSkillLink 安全删除路径，绝不递归删除源目录内容
//   - os.Lstat 检测是否为 symlink/junction
//   - 是链接 → os.Remove（只删链接节点）
//   - 不是链接（旧版复制目录）→ os.RemoveAll（正常递归删除）
func RemoveSkillLink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		// 是链接（symlink 或 junction），只删链接节点
		return os.Remove(path)
	}

	// 普通目录（旧版复制），正常递归删除
	return os.RemoveAll(path)
}
