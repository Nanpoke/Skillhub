package tools

// Adapter 定义 AI 工具适配器接口
type Adapter interface {
	// ID 返回工具的唯一标识符
	ID() string

	// Name 返回工具的显示名称
	Name() string

	// SkillsPath 返回该工具的 Skills 目录路径
	SkillsPath() string

	// IsInstalled 检查工具是否已安装
	IsInstalled() bool

	// EnableSkill 启用指定 Skill（复制到工具目录）
	EnableSkill(skillName string, skillPath string) error

	// DisableSkill 禁用指定 Skill（从工具目录删除）
	DisableSkill(skillName string) error

	// IsSkillEnabled 检查 Skill 是否已在该工具中启用
	IsSkillEnabled(skillName string) bool

	// ListEnabledSkills 列出该工具中已启用的所有 Skills
	ListEnabledSkills() ([]string, error)
}
