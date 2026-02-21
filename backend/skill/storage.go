package skill

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	SkillsDirName   = "skills"
	GitDirName      = "git"
	MetadataDirName = "metadata"
	HistoryDirName  = "history"
	SettingsFile    = "settings.json"
)

// Storage 处理文件系统存储操作
type Storage struct {
	basePath string
}

// NewStorage 创建新的存储实例
func NewStorage(basePath string) *Storage {
	return &Storage{basePath: basePath}
}

// GetBasePath 获取基础路径
func (s *Storage) GetBasePath() string {
	return s.basePath
}

// Initialize 初始化存储目录结构
func (s *Storage) Initialize() error {
	dirs := []string{
		s.SkillsPath(),
		s.GitPath(),
		s.MetadataPath(),
		s.HistoryPath(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// SkillsPath 返回 Skills 存储路径
func (s *Storage) SkillsPath() string {
	return filepath.Join(s.basePath, SkillsDirName)
}

// GitPath 返回 Git 目录路径
func (s *Storage) GitPath() string {
	return filepath.Join(s.basePath, GitDirName)
}

// MetadataPath 返回元数据目录路径
func (s *Storage) MetadataPath() string {
	return filepath.Join(s.basePath, MetadataDirName)
}

// HistoryPath 返回历史记录目录路径
func (s *Storage) HistoryPath() string {
	return filepath.Join(s.basePath, HistoryDirName)
}

// SettingsFilePath 返回设置文件路径
func (s *Storage) SettingsFilePath() string {
	return filepath.Join(s.basePath, SettingsFile)
}

// SkillPath 返回指定 Skill 的路径
func (s *Storage) SkillPath(name string) string {
	return filepath.Join(s.SkillsPath(), name)
}

// SkillGitPath 返回指定 Skill 的 Git 目录路径
func (s *Storage) SkillGitPath(name string) string {
	return filepath.Join(s.GitPath(), name)
}

// SkillMetadataPath 返回指定 Skill 的元数据文件路径
func (s *Storage) SkillMetadataPath(name string) string {
	return filepath.Join(s.MetadataPath(), name+".json")
}

// SkillExists 检查 Skill 是否存在
func (s *Storage) SkillExists(name string) bool {
	_, err := os.Stat(s.SkillPath(name))
	return err == nil
}

// ListSkills 列出所有已安装的 Skills
func (s *Storage) ListSkills() ([]string, error) {
	var skills []string

	entries, err := os.ReadDir(s.SkillsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return skills, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			skills = append(skills, entry.Name())
		}
	}

	return skills, nil
}

// SaveMetadata 保存 Skill 元数据
func (s *Storage) SaveMetadata(name string, meta *Metadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	path := s.SkillMetadataPath(name)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// LoadMetadata 加载 Skill 元数据
func (s *Storage) LoadMetadata(name string) (*Metadata, error) {
	path := s.SkillMetadataPath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	var meta Metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &meta, nil
}

// DeleteMetadata 删除 Skill 元数据
func (s *Storage) DeleteMetadata(name string) error {
	path := s.SkillMetadataPath(name)
	return os.Remove(path)
}

// LoadAllSkillsMetadata 加载所有 Skill 的元数据
func (s *Storage) LoadAllSkillsMetadata() ([]*Metadata, error) {
	names, err := s.ListSkills()
	if err != nil {
		return nil, err
	}

	var metas []*Metadata
	for _, name := range names {
		meta, err := s.LoadMetadata(name)
		if err != nil {
			continue // 跳过无法加载的
		}
		metas = append(metas, meta)
	}

	return metas, nil
}

// LoadSettings 加载应用设置
func (s *Storage) LoadSettings() (*AppSettings, error) {
	path := s.SettingsFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 返回默认设置
			return &AppSettings{
				FirstRun:        true,
				Theme:          "system",
				AutoUpdateCheck: true,
				UpdateFrequency: "daily",
			}, nil
		}
		return nil, err
	}

	var settings AppSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// SaveSettings 保存应用设置
func (s *Storage) SaveSettings(settings *AppSettings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.SettingsFilePath(), data, 0644)
}

// AppendOperationLog 追加操作日志
func (s *Storage) AppendOperationLog(log *OperationLog) error {
	if err := os.MkdirAll(s.HistoryPath(), 0755); err != nil {
		return err
	}

	path := filepath.Join(s.HistoryPath(), "operations.log")

	// 以追加模式打开文件
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// 每行一个 JSON
	data, err := json.Marshal(log)
	if err != nil {
		return err
	}

	_, err = f.WriteString(string(data) + "\n")
	return err
}

// ReadOperationLogs 读取操作日志
func (s *Storage) ReadOperationLogs(since time.Time) ([]OperationLog, error) {
	path := filepath.Join(s.HistoryPath(), "operations.log")

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var logs []OperationLog
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var log OperationLog
		if err := json.Unmarshal([]byte(line), &log); err != nil {
			continue
		}

		// 只返回指定时间之后的日志
		if log.Timestamp.After(since) {
			logs = append(logs, log)
		}
	}

	// 反转日志顺序，使最新的日志在前面
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}

	return logs, scanner.Err()
}

// CleanOldLogs 清理超过 10 天的日志
func (s *Storage) CleanOldLogs() error {
	tenDaysAgo := time.Now().AddDate(0, 0, -10)

	logs, err := s.ReadOperationLogs(time.Time{}) // 读取所有日志
	if err != nil {
		return err
	}

	// 过滤保留 10 天内的日志
	var recentLogs []OperationLog
	for _, log := range logs {
		if log.Timestamp.After(tenDaysAgo) {
			recentLogs = append(recentLogs, log)
		}
	}

	// 重写日志文件
	path := filepath.Join(s.HistoryPath(), "operations.log")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, log := range recentLogs {
		data, _ := json.Marshal(log)
		f.WriteString(string(data) + "\n")
	}

	return nil
}

// DeleteSkill 删除 Skill（包括元数据和 Git 目录）
func (s *Storage) DeleteSkill(name string) error {
	// 删除 Skill 目录
	if err := os.RemoveAll(s.SkillPath(name)); err != nil {
		return err
	}

	// 删除 Git 目录
	os.RemoveAll(s.SkillGitPath(name))

	// 删除元数据
	os.Remove(s.SkillMetadataPath(name))

	return nil
}

// CustomToolsFilePath 返回自定义工具配置文件路径
func (s *Storage) CustomToolsFilePath() string {
	return filepath.Join(s.basePath, "custom-tools.json")
}

// LoadCustomTools 加载自定义工具列表
func (s *Storage) LoadCustomTools() ([]*CustomTool, error) {
	path := s.CustomToolsFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []*CustomTool{}, nil
		}
		return nil, err
	}

	var result struct {
		Tools []*CustomTool `json:"tools"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result.Tools, nil
}

// SaveCustomTools 保存自定义工具列表
func (s *Storage) SaveCustomTools(tools []*CustomTool) error {
	data, err := json.MarshalIndent(map[string]interface{}{
		"tools": tools,
	}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.CustomToolsFilePath(), data, 0644)
}

// AddCustomTool 添加一个自定义工具
func (s *Storage) AddCustomTool(tool *CustomTool) error {
	tools, err := s.LoadCustomTools()
	if err != nil {
		return err
	}

	// 检查是否已存在
	for _, t := range tools {
		if t.ID == tool.ID {
			return fmt.Errorf("tool with id %s already exists", tool.ID)
		}
	}

	tools = append(tools, tool)
	return s.SaveCustomTools(tools)
}

// RemoveCustomTool 删除一个自定义工具
func (s *Storage) RemoveCustomTool(id string) error {
	tools, err := s.LoadCustomTools()
	if err != nil {
		return err
	}

	// 过滤掉要删除的工具
	var newTools []*CustomTool
	for _, t := range tools {
		if t.ID != id {
			newTools = append(newTools, t)
		}
	}

	return s.SaveCustomTools(newTools)
}
