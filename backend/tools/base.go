package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skillhub/backend/utils"
)

// validateSkillName 验证 skill 名称安全性
func validateSkillName(name string) error {
	if name == "" {
		return fmt.Errorf("skill name cannot be empty")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("invalid skill name: contains '..'")
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("invalid skill name: contains path separator")
	}
	return nil
}

// BaseAdapter 提供工具适配器的公共实现
type BaseAdapter struct {
	id         string
	name       string
	skillsPath string
}

// NewBaseAdapter 创建基础适配器
func NewBaseAdapter(id, name, skillsPath string) *BaseAdapter {
	return &BaseAdapter{
		id:         id,
		name:       name,
		skillsPath: skillsPath,
	}
}

func (a *BaseAdapter) ID() string {
	return a.id
}

func (a *BaseAdapter) Name() string {
	return a.name
}

func (a *BaseAdapter) SkillsPath() string {
	return a.skillsPath
}

func (a *BaseAdapter) IsInstalled() bool {
	// 检查工具目录是否存在
	toolDir := filepath.Dir(a.skillsPath)
	_, err := os.Stat(toolDir)
	return err == nil
}

func (a *BaseAdapter) EnableSkill(skillName string, skillPath string) error {
	// 验证 skillName 安全性
	if err := validateSkillName(skillName); err != nil {
		return err
	}

	// 验证 skillPath 不逃逸
	if err := utils.ValidatePathInDir(skillPath, skillPath); err != nil {
		return fmt.Errorf("invalid skill path: %w", err)
	}

	// 确保 Skills 目录存在
	if err := os.MkdirAll(a.skillsPath, 0755); err != nil {
		return err
	}

	// 目标路径
	destPath := filepath.Join(a.skillsPath, skillName)

	// 如果已存在，先删除
	if _, err := os.Stat(destPath); err == nil {
		if err := os.RemoveAll(destPath); err != nil {
			return err
		}
	}

	// 复制整个目录（不跳过 .git）
	return utils.CopyDir(skillPath, destPath, false)
}

func (a *BaseAdapter) DisableSkill(skillName string) error {
	// 验证 skillName 安全性
	if err := validateSkillName(skillName); err != nil {
		return err
	}

	destPath := filepath.Join(a.skillsPath, skillName)
	return os.RemoveAll(destPath)
}

func (a *BaseAdapter) IsSkillEnabled(skillName string) bool {
	destPath := filepath.Join(a.skillsPath, skillName)
	_, err := os.Stat(destPath)
	return err == nil
}

func (a *BaseAdapter) ListEnabledSkills() ([]string, error) {
	var skills []string

	// 检查目录是否存在
	if _, err := os.Stat(a.skillsPath); os.IsNotExist(err) {
		return skills, nil
	}

	// 读取目录内容
	entries, err := os.ReadDir(a.skillsPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			skills = append(skills, entry.Name())
		}
	}

	return skills, nil
}
