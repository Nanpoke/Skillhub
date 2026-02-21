package skill

import "time"

// SourceType 定义 Skill 的来源类型
type SourceType string

const (
	SourceTypeGit   SourceType = "git"
	SourceTypeLocal SourceType = "local"
)

// Skill 表示一个 Skill 的完整信息
type Skill struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	OriginalName  string            `json:"original_name"`
	Author        string            `json:"author"`
	Description   string            `json:"description"`
	SourceType    SourceType        `json:"source_type"`
	SourceURL     string            `json:"source_url"`
	Category      string            `json:"category"`
	Tags          []string          `json:"tags"`
	Notes         string            `json:"notes"`
	ToolsEnabled  map[string]bool   `json:"tools_enabled"`
	InstalledAt   time.Time         `json:"installed_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	HasUpdate     bool              `json:"has_update"`
}

// Metadata 用于持久化存储的元数据
type Metadata struct {
	Name         string      `json:"name"`
	OriginalName string      `json:"original_name"`
	Author       string      `json:"author"`
	SourceType   SourceType  `json:"source_type"`
	SourceURL    string      `json:"source_url"`
	Category     string      `json:"category"`
	Tags         []string    `json:"tags"`
	Notes        string      `json:"notes"`
	InstalledAt  time.Time   `json:"installed_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	GitVersion   string      `json:"git_version"` // 当前 Git 版本（tag）
	HasUpdate    bool        `json:"has_update"`  // 是否有更新可用
}

// InstallOptions 安装选项
type InstallOptions struct {
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Notes    string   `json:"notes"`
}

// SkillInfo 用于安装前预览的 Skill 信息
type SkillInfo struct {
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

// OperationLog 操作日志条目
type OperationLog struct {
	Timestamp  time.Time      `json:"timestamp"`
	Action     string         `json:"action"`
	Source     string         `json:"source"`
	Skills     []SkillStatus  `json:"skills"`
	DurationMs int64          `json:"duration_ms"`
}

// SkillStatus 技能安装状态
type SkillStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// AppSettings 应用设置
type AppSettings struct {
	SkillHubPath     string   `json:"skillhub_path"`
	Theme            string   `json:"theme"`
	AutoUpdateCheck  bool     `json:"auto_update_check"`
	UpdateFrequency  string   `json:"update_frequency"`
	FirstRun         bool     `json:"first_run"`
	CustomCategories []string `json:"custom_categories"` // 用户自定义分类
}

// CategoryInfo 分类信息
type CategoryInfo struct {
	Name     string `json:"name"`
	IsPreset bool   `json:"is_preset"`
}

// ToolInfo 工具信息
type ToolInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	SkillsPath  string `json:"skills_path"`
	IsInstalled bool   `json:"is_installed"`
	IsEnabled   bool   `json:"is_enabled"` // 工具是否启用（从配置文件读取）
}

// CustomTool 用户自定义添加的工具
type CustomTool struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	SkillsPath string `json:"skills_path"`
	Enabled    bool   `json:"enabled"`
	DateAdded  string `json:"date_added"`
}

// SkillUpdateInfo 单个 Skill 的更新信息
type SkillUpdateInfo struct {
	Name           string `json:"name"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
}
