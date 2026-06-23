//go:build !windows

package utils

import (
	"os"
)

func createSkillLink(linkPath, targetPath string) error {
	return os.Symlink(targetPath, linkPath)
}

func isSkillLink(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.Mode()&os.ModeSymlink != 0, nil
}
