//go:build windows

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func createSkillLink(linkPath, targetPath string) error {
	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for target: %w", err)
	}
	linkAbs, err := filepath.Abs(linkPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for link: %w", err)
	}

	cmd := exec.Command("cmd", "/c", "mklink", "/J", linkAbs, targetAbs)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mklink /J failed: %w, output: %s", err, string(output))
	}
	return nil
}

func isSkillLink(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	// Windows 上 junction 表现为 ModeSymlink
	return info.Mode()&os.ModeSymlink != 0, nil
}
