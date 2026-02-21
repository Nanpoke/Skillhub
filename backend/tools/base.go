package tools

import (
	"os"
	"path/filepath"
)

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

	// 复制整个目录
	return copyDir(skillPath, destPath)
}

func (a *BaseAdapter) DisableSkill(skillName string) error {
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

// copyDir 递归复制目录
func copyDir(src, dst string) error {
	// 获取源目录信息
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 创建目标目录
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	// 读取源目录
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	// 读取源文件
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// 获取源文件权限
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 写入目标文件
	return os.WriteFile(dst, data, info.Mode())
}
