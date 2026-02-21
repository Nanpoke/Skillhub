package skill

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	stdruntime "runtime"
	"strings"
	"time"

	"skillhub/backend/tools"
	"skillhub/backend/utils"
)

// PresetCategories 预设分类列表
var PresetCategories = []string{
	"内容创作", "认知增强", "开发辅助",
	"数据分析", "教育学习", "AI/LLM", "其他",
}

// Manager 管理 Skills 的核心逻辑
type Manager struct {
	storage  *Storage
	registry *tools.Registry
}

// NewManager 创建新的 Skill 管理器
func NewManager(basePath string) *Manager {
	return &Manager{
		storage:  NewStorage(basePath),
		registry: tools.NewRegistry(),
	}
}

// Initialize 初始化管理器
func (m *Manager) Initialize() error {
	if err := m.storage.Initialize(); err != nil {
		return err
	}
	// 注册自定义工具适配器
	m.registerCustomToolAdapters()
	return nil
}

// registerCustomToolAdapters 加载并注册自定义工具适配器
func (m *Manager) registerCustomToolAdapters() {
	customTools, err := m.storage.LoadCustomTools()
	if err != nil {
		return
	}

	for _, ct := range customTools {
		if ct.Enabled {
			adapter := tools.NewBaseAdapter(ct.ID, ct.Name, ct.SkillsPath)
			m.registry.Register(adapter)
		}
	}
}

// GetStorage 获取存储实例
func (m *Manager) GetStorage() *Storage {
	return m.storage
}

// GetRegistry 获取工具注册表
func (m *Manager) GetRegistry() *tools.Registry {
	return m.registry
}

// ListSkills 列出所有 Skills 及其状态
func (m *Manager) ListSkills() ([]Skill, error) {
	names, err := m.storage.ListSkills()
	if err != nil {
		return nil, err
	}

	var skills []Skill
	for _, name := range names {
		skill, err := m.GetSkill(name)
		if err != nil {
			// 跳过无法加载的 Skill
			continue
		}
		skills = append(skills, *skill)
	}

	return skills, nil
}

// GetSkill 获取单个 Skill 的完整信息
func (m *Manager) GetSkill(name string) (*Skill, error) {
	// 加载元数据
	meta, err := m.storage.LoadMetadata(name)
	if err != nil {
		return nil, err
	}

	// 构建工具启用状态
	toolsEnabled := make(map[string]bool)
	for _, adapter := range m.registry.GetAll() {
		toolsEnabled[adapter.ID()] = adapter.IsSkillEnabled(name)
	}

	return &Skill{
		ID:           name,
		Name:         meta.Name,
		OriginalName: meta.OriginalName,
		Author:       meta.Author,
		Description:  "", // 从 SKILL.md 读取，这里暂时留空
		SourceType:   meta.SourceType,
		SourceURL:    meta.SourceURL,
		Category:     meta.Category,
		Tags:         meta.Tags,
		Notes:        meta.Notes,
		ToolsEnabled: toolsEnabled,
		InstalledAt:  meta.InstalledAt,
		UpdatedAt:    meta.UpdatedAt,
		HasUpdate:    meta.HasUpdate, // 从元数据读取更新状态
	}, nil
}

// ToggleSkill 切换 Skill 在指定工具中的启用状态
func (m *Manager) ToggleSkill(skillName string, toolID string) (bool, error) {
	adapter := m.registry.Get(toolID)
	if adapter == nil {
		return false, fmt.Errorf("unknown tool: %s", toolID)
	}

	skillPath := m.storage.SkillPath(skillName)
	if _, err := os.Stat(skillPath); err != nil {
		return false, fmt.Errorf("skill not found: %s", skillName)
	}

	isEnabled := adapter.IsSkillEnabled(skillName)

	var err error
	if isEnabled {
		err = adapter.DisableSkill(skillName)
	} else {
		err = adapter.EnableSkill(skillName, skillPath)
	}

	if err != nil {
		return false, err
	}

	// 记录操作
	action := "enable"
	if isEnabled {
		action = "disable"
	}
	m.logOperation(action, skillName, "success")

	return !isEnabled, nil
}

// DeleteSkill 删除 Skill
func (m *Manager) DeleteSkill(name string) error {
	// 先从所有工具中禁用
	for _, adapter := range m.registry.GetAll() {
		if adapter.IsSkillEnabled(name) {
			adapter.DisableSkill(name)
		}
	}

	// 删除 Skill 文件
	if err := m.storage.DeleteSkill(name); err != nil {
		return err
	}

	m.logOperation("delete", name, "success")
	return nil
}

// UpdateMetadata 更新 Skill 元数据
func (m *Manager) UpdateMetadata(skillID string, category string, tags []string, notes string) error {
	// 加载现有元数据
	meta, err := m.storage.LoadMetadata(skillID)
	if err != nil {
		return err
	}

	// 更新字段
	meta.Category = category
	meta.Tags = tags
	meta.Notes = notes
	meta.UpdatedAt = time.Now()

	// 保存
	if err := m.storage.SaveMetadata(skillID, meta); err != nil {
		return err
	}

	m.logOperation("update", skillID, "success")
	return nil
}

// InstallFromPath 从本地路径安装 Skill
func (m *Manager) InstallFromPath(srcPath string, options InstallOptions) (*Skill, error) {
	// 扫描 SKILL.md
	skillInfo, err := m.ScanSkillPath(srcPath)
	if err != nil {
		return nil, err
	}

	// 生成名称
	name := fmt.Sprintf("%s-%s", skillInfo.Author, skillInfo.Name)

	// 检查是否已存在
	if m.storage.SkillExists(name) {
		return nil, fmt.Errorf("skill already exists: %s", name)
	}

	// 复制 Skill 文件
	destPath := m.storage.SkillPath(name)
	if err := copyDir(srcPath, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy skill: %w", err)
	}

	// 保存元数据
	now := time.Now()
	meta := &Metadata{
		Name:         name,
		OriginalName: skillInfo.Name,
		Author:       skillInfo.Author,
		SourceType:   SourceTypeLocal,
		SourceURL:    srcPath,
		Category:     options.Category,
		Tags:         options.Tags,
		Notes:        options.Notes,
		InstalledAt:  now,
		UpdatedAt:    now,
	}

	if err := m.storage.SaveMetadata(name, meta); err != nil {
		// 回滚
		os.RemoveAll(destPath)
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	m.logOperation("install", name, "success")

	return m.GetSkill(name)
}

// LocalScanResult 本地扫描结果
type LocalScanResult struct {
	TempPath string       // 临时目录（如果是压缩包则有值）
	IsZip    bool         // 是否是压缩包
	SourcePath string     // 原始路径
	Skills   []*SkillInfo // 发现的 Skills
}

// ScanLocalPath 扫描本地路径（文件夹或压缩包）
func (m *Manager) ScanLocalPath(path string) (*LocalScanResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("路径不存在: %s", path)
	}

	// 判断是文件还是目录
	if info.IsDir() {
		// 文件夹：直接扫描
		return m.scanLocalFolder(path, "")
	}

	// 文件：检查是否是压缩包
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".zip" || ext == ".skill" {
		return m.scanLocalZip(path)
	}

	return nil, fmt.Errorf("不支持的文件格式: %s", ext)
}

// scanLocalFolder 扫描本地文件夹
func (m *Manager) scanLocalFolder(folderPath string, tempPath string) (*LocalScanResult, error) {
	var skills []*SkillInfo

	// 首先检查根目录是否有 SKILL.md
	rootInfo, err := m.ScanSkillPath(folderPath)
	if err == nil {
		rootInfo.Path = ""
		skills = append(skills, rootInfo)
	}

	// 递归扫描子目录
	m.scanSkillsRecursive(folderPath, "", &skills)

	if len(skills) == 0 {
		return nil, fmt.Errorf("未找到有效的 Skill（缺少 SKILL.md 文件）")
	}

	return &LocalScanResult{
		TempPath:   tempPath,
		IsZip:      tempPath != "",
		SourcePath: folderPath,
		Skills:     skills,
	}, nil
}

// scanSkillsRecursive 递归扫描目录中的 Skills
func (m *Manager) scanSkillsRecursive(basePath string, relativePath string, skills *[]*SkillInfo) {
	fullPath := basePath
	if relativePath != "" {
		fullPath = filepath.Join(basePath, relativePath)
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subRelativePath := entry.Name()
			if relativePath != "" {
				subRelativePath = filepath.Join(relativePath, entry.Name())
			}

			subFullPath := filepath.Join(basePath, subRelativePath)
			skillInfo, err := m.ScanSkillPath(subFullPath)
			if err == nil {
				// 找到 Skill，添加到列表
				skillInfo.Path = subRelativePath
				*skills = append(*skills, skillInfo)
			} else {
				// 不是 Skill 目录，继续递归扫描
				m.scanSkillsRecursive(basePath, subRelativePath, skills)
			}
		}
	}
}

// unzipFile 解压 ZIP 文件到指定目录（使用安全验证）
func unzipFile(zipPath, destDir string) error {
	return utils.UnzipFile(zipPath, destDir)
}

// scanLocalZip 扫描本地压缩包
func (m *Manager) scanLocalZip(zipPath string) (*LocalScanResult, error) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "skillhub-local-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 解压
	if err := unzipFile(zipPath, tempDir); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("解压失败: %w", err)
	}

	// 扫描解压后的目录
	result, err := m.scanLocalFolder(tempDir, tempDir)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, err
	}

	result.IsZip = true
	result.SourcePath = zipPath
	return result, nil
}

// InstallFromLocalTemp 从本地临时目录安装 Skill
func (m *Manager) InstallFromLocalTemp(tempPath string, skillRelativePath string, sourcePath string, options InstallOptions) (*Skill, error) {
	// 构建完整路径
	var skillPath string
	if skillRelativePath == "" {
		skillPath = tempPath
	} else {
		skillPath = filepath.Join(tempPath, skillRelativePath)
	}

	// 扫描 SKILL.md
	skillInfo, err := m.ScanSkillPath(skillPath)
	if err != nil {
		return nil, err
	}

	// 生成名称
	name := fmt.Sprintf("%s-%s", skillInfo.Author, skillInfo.Name)

	// 检查是否已存在
	if m.storage.SkillExists(name) {
		return nil, fmt.Errorf("Skill 已存在: %s", name)
	}

	// 复制 Skill 文件
	destPath := m.storage.SkillPath(name)
	if err := copyDir(skillPath, destPath); err != nil {
		return nil, fmt.Errorf("复制 Skill 失败: %w", err)
	}

	// 保存元数据
	now := time.Now()
	meta := &Metadata{
		Name:         name,
		OriginalName: skillInfo.Name,
		Author:       skillInfo.Author,
		SourceType:   SourceTypeLocal,
		SourceURL:    sourcePath,
		Category:     options.Category,
		Tags:         options.Tags,
		Notes:        options.Notes,
		InstalledAt:  now,
		UpdatedAt:    now,
	}

	if err := m.storage.SaveMetadata(name, meta); err != nil {
		os.RemoveAll(destPath)
		return nil, fmt.Errorf("保存元数据失败: %w", err)
	}

	m.logOperation("install", name, "success")
	return m.GetSkill(name)
}

// CleanupLocalTemp 清理本地临时目录
func (m *Manager) CleanupLocalTemp(tempPath string) {
	if tempPath != "" {
		os.RemoveAll(tempPath)
	}
}

// GitInstallResult Git 安装结果
type GitInstallResult struct {
	TempPath string       // 临时目录
	GitURL   string       // Git 仓库 URL
	Skills   []*SkillInfo // 发现的 Skills
}

// CloneFromGit 从 Git 仓库克隆并扫描 Skills
func (m *Manager) CloneFromGit(url string) (*GitInstallResult, error) {
	git := utils.NewGitClient()

	if !git.IsGitInstalled() {
		return nil, fmt.Errorf("Git 未安装，请先安装 Git")
	}

	// 解析 Git URL 获取 owner/repo
	gitInfo, err := git.ParseGitURL(url)
	if err != nil {
		return nil, err
	}

	// 克隆仓库
	result, err := git.Clone(gitInfo.FullURL)
	if err != nil {
		return nil, err
	}

	// 扫描 Skills
	var skills []*SkillInfo
	for _, skillPath := range result.Skills {
		fullPath := filepath.Join(result.TempPath, skillPath)
		info, err := m.ScanSkillPath(fullPath)
		if err == nil {
			info.Path = skillPath // 保存相对路径
			skills = append(skills, info)
		}
	}

	// 如果根目录没有找到 Skill，尝试扫描整个目录
	if len(skills) == 0 {
		// 检查根目录
		info, err := m.ScanSkillPath(result.TempPath)
		if err == nil {
			info.Path = ""
			skills = append(skills, info)
		}
	}

	return &GitInstallResult{
		TempPath: result.TempPath,
		GitURL:   gitInfo.FullURL, // 返回 Git URL
		Skills:   skills,
	}, nil
}

// InstallFromGit 从 Git 临时目录安装 Skill
func (m *Manager) InstallFromGit(tempPath string, skillRelativePath string, gitURL string, options InstallOptions) (*Skill, error) {
	// 构建完整路径
	skillPath := filepath.Join(tempPath, skillRelativePath)

	// 扫描 SKILL.md
	skillInfo, err := m.ScanSkillPath(skillPath)
	if err != nil {
		return nil, err
	}

	// 解析 Git URL 获取作者信息（优先级最高，覆盖从 SKILL.md/LICENSE 解析的作者）
	git := utils.NewGitClient()
	gitInfo, err := git.ParseGitURL(gitURL)
	if err == nil {
		// 使用 Git URL 的作者信息
		skillInfo.Author = gitInfo.Owner
	}

	// 生成名称
	name := fmt.Sprintf("%s-%s", skillInfo.Author, skillInfo.Name)

	// 检查是否已存在
	if m.storage.SkillExists(name) {
		return nil, fmt.Errorf("skill already exists: %s", name)
	}

	// 复制 Skill 文件（不包含 .git）
	destPath := m.storage.SkillPath(name)
	if err := copyDir(skillPath, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy skill: %w", err)
	}

	// 移动 .git 目录到集中管理
	gitSrcPath := tempPath
	if skillRelativePath != "" {
		// 如果 Skill 在子目录，.git 在仓库根目录
		gitSrcPath = tempPath
	}
	gitDestPath := m.storage.SkillGitPath(name)

	// 检查并移动 .git 目录
	if err := git.MoveGitDir(gitSrcPath, gitDestPath); err != nil {
		// 清理已复制的文件
		os.RemoveAll(destPath)
		return nil, fmt.Errorf("failed to move .git directory: %w", err)
	}

	// 保存元数据（使用 skillInfo.Author，即 Git URL 的作者）
	now := time.Now()
	meta := &Metadata{
		Name:         name,
		OriginalName: skillInfo.Name,
		Author:       skillInfo.Author,
		SourceType:   SourceTypeGit,
		SourceURL:    gitURL,
		Category:     options.Category,
		Tags:         options.Tags,
		Notes:        options.Notes,
		InstalledAt:  now,
		UpdatedAt:    now,
	}

	if err := m.storage.SaveMetadata(name, meta); err != nil {
		// 回滚
		os.RemoveAll(destPath)
		os.RemoveAll(gitDestPath)
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	m.logOperation("install", name, "success")

	return m.GetSkill(name)
}

// CleanupClone 清理克隆的临时目录
func (m *Manager) CleanupClone(tempPath string) {
	if tempPath != "" {
		os.RemoveAll(tempPath)
	}
}

// ParseGitURL 解析 Git URL
func (m *Manager) ParseGitURL(url string) (*utils.GitURLInfo, error) {
	git := utils.NewGitClient()
	return git.ParseGitURL(url)
}

// ScanSkillPath 扫描路径获取 Skill 信息
func (m *Manager) ScanSkillPath(path string) (*SkillInfo, error) {
	// 检查 SKILL.md 是否存在
	skillFile := filepath.Join(path, "SKILL.md")
	if _, err := os.Stat(skillFile); err != nil {
		return nil, fmt.Errorf("SKILL.md not found in %s", path)
	}

	// 读取 SKILL.md
	content, err := os.ReadFile(skillFile)
	if err != nil {
		return nil, err
	}

	// 解析 frontmatter（现在包含 author）
	name, description, author := parseFrontmatter(string(content))

	// 如果 SKILL.md 中没有 author，尝试从 LICENSE.txt 获取
	if author == "" {
		author = parseAuthorFromLicense(path)
	}

	// 如果仍然没有 author，从路径获取作者信息
	dirName := filepath.Base(path)
	if author == "" && strings.Contains(dirName, "-") {
		parts := strings.SplitN(dirName, "-", 2)
		if len(parts) == 2 {
			author = parts[0]
			if name == "" {
				name = parts[1]
			}
		}
	}

	// 最后，如果仍然没有 author 和 name，使用目录名
	if author == "" {
		author = "unknown"
	}
	if name == "" {
		name = dirName
	}

	return &SkillInfo{
		Name:        name,
		Author:      author,
		Description: description,
		Path:        path,
	}, nil
}

// GetDetectedTools 获取检测到的已安装工具
func (m *Manager) GetDetectedTools() []ToolInfo {
	var toolInfos []ToolInfo
	for _, adapter := range m.registry.GetAll() {
		// 读取工具配置文件获取 enabled 状态
		toolID := adapter.ID()
		configPath := filepath.Join(m.storage.GetBasePath(), "config", toolID+".json")

		isEnabled := false // 默认未启用
		if data, err := os.ReadFile(configPath); err == nil {
			var config map[string]interface{}
			if err := json.Unmarshal(data, &config); err == nil {
				if enabled, ok := config["enabled"].(bool); ok {
					isEnabled = enabled
				}
			}
		}

		toolInfos = append(toolInfos, ToolInfo{
			ID:          adapter.ID(),
			Name:        adapter.Name(),
			SkillsPath:  adapter.SkillsPath(),
			IsInstalled: adapter.IsInstalled(),
			IsEnabled:   isEnabled, // 从配置文件读取
		})
	}
	return toolInfos
}

// GetSettings 获取应用设置
func (m *Manager) GetSettings() (*AppSettings, error) {
	return m.storage.LoadSettings()
}

// SaveSettings 保存应用设置
func (m *Manager) SaveSettings(settings *AppSettings) error {
	return m.storage.SaveSettings(settings)
}

// InitializeWizard 初始化向导配置
// storagePath: 用户选择的存储路径
// selectedTools: 用户选择的预置工具 ID 列表
// customTools: 用户添加的自定义工具列表
func (m *Manager) InitializeWizard(storagePath string, selectedTools []string, customTools []*CustomTool) error {
	// 1. 展开路径中的 ~ 符号
	expandedPath := utils.ExpandPath(storagePath)

	// 2. 创建新的存储实例
	newStorage := NewStorage(expandedPath)

	// 3. 初始化目录结构
	if err := newStorage.Initialize(); err != nil {
		return fmt.Errorf("初始化目录结构失败: %w", err)
	}

	// 4. 保存工具配置
	// 获取所有检测到的预置工具
	allAdapters := m.registry.GetAll()

	// 为每个工具创建配置文件
	for _, adapter := range allAdapters {
		toolID := adapter.ID()
		configPath := filepath.Join(newStorage.GetBasePath(), "config", toolID+".json")

		// 检查工具是否被选中
		isSelected := false
		for _, id := range selectedTools {
			if id == toolID {
				isSelected = true
				break
			}
		}

		// 创建配置（根据是否选中决定 enabled 状态）
		config := map[string]interface{}{
			"enabled":       isSelected,
			"skills_path":   adapter.SkillsPath(),
			"tool_name":     adapter.Name(),
			"date_added":    time.Now().Format("2006-01-02"),
		}

		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("序列化配置失败: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			return fmt.Errorf("创建配置目录失败: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}
	}

	// 5. 保存自定义工具配置
	for _, customTool := range customTools {
		if err := newStorage.AddCustomTool(customTool); err != nil {
			return fmt.Errorf("保存自定义工具失败: %w", err)
		}
	}

	// 6. 保存应用设置
	settings := &AppSettings{
		SkillHubPath:    expandedPath,
		Theme:           "system",
		AutoUpdateCheck: true,
		UpdateFrequency: "daily",
		FirstRun:        false,  // 向导完成后标记为非首次运行
	}

	if err := newStorage.SaveSettings(settings); err != nil {
		return fmt.Errorf("保存设置失败: %w", err)
	}

	// 7. 同时在应用配置路径保存一份设置，确保下次启动时能找到
	// 这样即使用户设置了自定义路径，启动时也能从配置路径读取到正确的配置
	// 配置路径使用 %APPDATA%\SkillHub（Windows）或 ~/.config/skillhub（macOS/Linux）
	configPath := GetAppConfigPath()
	configStorage := NewStorage(configPath)
	if err := configStorage.Initialize(); err == nil {
		configStorage.SaveSettings(settings)
	}

	// 8. 更新当前存储
	m.storage = newStorage

	return nil
}

// GetAppConfigPath 获取应用配置存储路径（用于保存 settings.json）
// Windows: %APPDATA%\SkillHub\
// macOS/Linux: ~/.config/skillhub/
// 这个路径用于存储配置"指针"，指向用户设置的实际数据路径
func GetAppConfigPath() string {
	if stdruntime.GOOS == "windows" {
		// Windows: 使用 %APPDATA%
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		return filepath.Join(appData, "SkillHub")
	}
	// macOS/Linux: 使用 ~/.config/
	home := utils.GetHomeDir()
	return filepath.Join(home, ".config", "skillhub")
}

// GetCustomTools 获取自定义工具列表
func (m *Manager) GetCustomTools() ([]*CustomTool, error) {
	return m.storage.LoadCustomTools()
}

// RemoveCustomTool 删除自定义工具
func (m *Manager) RemoveCustomTool(id string) error {
	// 1. 获取自定义工具信息
	customTools, err := m.storage.LoadCustomTools()
	if err != nil {
		return err
	}

	var targetTool *CustomTool
	for _, t := range customTools {
		if t.ID == id {
			targetTool = t
			break
		}
	}

	if targetTool == nil {
		return fmt.Errorf("custom tool not found: %s", id)
	}

	// 2. 删除已同步到该工具目录的 Skill 文件
	skillNames, err := m.storage.ListSkills()
	if err != nil {
		return err
	}

	for _, skillName := range skillNames {
		skillDir := filepath.Join(targetTool.SkillsPath, skillName)
		if _, err := os.Stat(skillDir); err == nil {
			// Skill 文件存在，删除它
			if err := os.RemoveAll(skillDir); err != nil {
				// 记录错误但继续处理其他 Skill
				fmt.Printf("Warning: failed to remove skill %s from %s: %v\n", skillName, targetTool.SkillsPath, err)
			}
		}
	}

	// 3. 从 custom-tools.json 中删除
	return m.storage.RemoveCustomTool(id)
}

// ToggleToolEnabled 切换工具的启用状态
func (m *Manager) ToggleToolEnabled(toolID string) error {
	adapter := m.registry.Get(toolID)
	if adapter == nil {
		return fmt.Errorf("unknown tool: %s", toolID)
	}

	// 读取工具配置
	configPath := filepath.Join(m.storage.GetBasePath(), "config", toolID+".json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		// 配置文件不存在，创建新配置
		config := map[string]interface{}{
			"enabled": true,
		}
		jsonData, _ := json.MarshalIndent(config, "", "  ")
		os.MkdirAll(filepath.Dir(configPath), 0755)
		os.WriteFile(configPath, jsonData, 0644)
		return nil
	}

	var config map[string]interface{}
	json.Unmarshal(data, &config)

	// 切换启用状态
	if enabled, ok := config["enabled"].(bool); ok {
		config["enabled"] = !enabled
	}

	// 保存配置
	jsonData, _ := json.MarshalIndent(config, "", "  ")
	if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// AddCustomTool 添加自定义工具
func (m *Manager) AddCustomTool(tool *CustomTool) error {
	if err := m.storage.AddCustomTool(tool); err != nil {
		return err
	}
	// 注册新工具的适配器
	if tool.Enabled {
		adapter := tools.NewBaseAdapter(tool.ID, tool.Name, tool.SkillsPath)
		m.registry.Register(adapter)
	}
	return nil
}

// logOperation 记录操作日志
func (m *Manager) logOperation(action, skillName, status string) {
	log := &OperationLog{
		Timestamp: time.Now(),
		Action:    action,
		Source:    "local",
		Skills: []SkillStatus{
			{Name: skillName, Status: status},
		},
	}
	m.storage.AppendOperationLog(log)
}

// copyDir 递归复制目录（跳过 .git）
func copyDir(src, dst string) error {
	return utils.CopyDir(src, dst, true)
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	return utils.CopyFile(src, dst)
}

// parseFrontmatter 解析 SKILL.md 的 frontmatter
func parseFrontmatter(content string) (name, description, author string) {
	lines := strings.Split(content, "\n")
	inFrontmatter := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			break
		}

		if inFrontmatter {
			if strings.HasPrefix(line, "name:") {
				name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				name = strings.Trim(name, "\"'")
			}
			if strings.HasPrefix(line, "author:") {
				author = strings.TrimSpace(strings.TrimPrefix(line, "author:"))
				author = strings.Trim(author, "\"'")
			}
			if strings.HasPrefix(line, "description:") {
				description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				description = strings.Trim(description, "\"'")
			}
		}
	}

	return
}

// parseAuthorFromLicense 从 LICENSE.txt 解析作者信息
func parseAuthorFromLicense(path string) string {
	licenseFile := filepath.Join(path, "LICENSE")
	if _, err := os.Stat(licenseFile); err != nil {
		return ""
	}

	content, err := os.ReadFile(licenseFile)
	if err != nil {
		return ""
	}

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 尝试匹配 "Copyright (c) YEAR <author>" 格式
		if strings.Contains(line, "Copyright") {
			// 提取 copyright 信息
			parts := strings.Fields(line)
			for i, part := range parts {
				if strings.EqualFold(part, "by") && i+1 < len(parts) {
					return strings.Join(parts[i+1:], " ")
				}
			}
		}
	}

	return ""
}

// CheckSkillUpdates 检查所有 Git 类型 Skills 的更新
func (m *Manager) CheckSkillUpdates() ([]SkillUpdateInfo, error) {
	skills, err := m.ListSkills()
	if err != nil {
		return nil, err
	}

	var updates []SkillUpdateInfo
	git := utils.NewGitClient()

	for _, skillInfo := range skills {
		// 只检查 Git 类型的 Skill
		if skillInfo.SourceType != SourceTypeGit {
			continue
		}

		// 解析 Git URL 获取 owner/repo
		gitInfo, err := git.ParseGitURL(skillInfo.SourceURL)
		if err != nil {
			continue
		}

		// 获取 Git 目录路径
		gitDir := m.storage.SkillGitPath(skillInfo.ID)
		if _, err := os.Stat(gitDir); err != nil {
			continue // Git 目录不存在，跳过
		}

		// 获取本地版本（使用支持分离 .git 目录的方法）
		localTag, err := git.GetTagWithGitDir(gitDir)
		if err != nil {
			continue
		}

		// 获取远程最新版本
		release, err := git.FetchLatestRelease(gitInfo.Owner, gitInfo.Repo)
		if err != nil {
			continue
		}

		// 比较版本
		hasUpdate := git.CompareVersions(localTag, release.TagName) < 0

		updates = append(updates, SkillUpdateInfo{
			Name:           skillInfo.ID,
			CurrentVersion: localTag,
			LatestVersion:  release.TagName,
			HasUpdate:      hasUpdate,
		})

		// 更新 HasUpdate 标识到元数据
		meta, _ := m.storage.LoadMetadata(skillInfo.ID)
		if meta != nil && meta.HasUpdate != hasUpdate {
			meta.HasUpdate = hasUpdate
			m.storage.SaveMetadata(skillInfo.ID, meta)
		}
	}

	return updates, nil
}

// UpdateSkill 更新单个 Skill
func (m *Manager) UpdateSkill(name string) error {
	// 检查 Skill 是否存在
	meta, err := m.storage.LoadMetadata(name)
	if err != nil {
		return fmt.Errorf("skill not found: %s", name)
	}

	// 只支持 Git 类型的 Skill
	if meta.SourceType != SourceTypeGit {
		return fmt.Errorf("只有 Git 仓库安装的 Skill 才能更新")
	}

	gitDir := m.storage.SkillGitPath(name)
	skillDir := m.storage.SkillPath(name)

	// 检查 Git 目录是否存在
	if _, err := os.Stat(gitDir); err != nil {
		return fmt.Errorf("Git 目录不存在，无法更新。此 Skill 可能是本地安装或 Git 信息丢失，请尝试重新安装")
	}

	// 执行 git pull 更新（使用分离的 .git 目录和工作目录）
	git := utils.NewGitClient()
	if err := git.PullWithGitDir(gitDir, skillDir); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	// 获取新的版本号
	newTag, err := git.GetTagWithGitDir(gitDir)
	if err == nil {
		meta.GitVersion = newTag
	}
	meta.UpdatedAt = time.Now()
	meta.HasUpdate = false // 更新后清除标识

	// 保存更新后的元数据
	if err := m.storage.SaveMetadata(name, meta); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	m.logOperation("update", name, "success")
	return nil
}

// copySkillFiles 复制 Skill 文件（排除 .git）
func (m *Manager) copySkillFiles(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// 跳过 .git 目录
		if entry.Name() == ".git" {
			continue
		}

		if entry.IsDir() {
			if err := m.copySkillFiles(srcPath, dstPath); err != nil {
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

// GetAllCategories 获取所有分类（预设 + 自定义）
func (m *Manager) GetAllCategories() ([]CategoryInfo, error) {
	var categories []CategoryInfo

	// 添加预设分类
	for _, name := range PresetCategories {
		categories = append(categories, CategoryInfo{
			Name:     name,
			IsPreset: true,
		})
	}

	// 加载自定义分类
	settings, err := m.storage.LoadSettings()
	if err != nil {
		return categories, nil // 返回预设分类
	}

	for _, name := range settings.CustomCategories {
		categories = append(categories, CategoryInfo{
			Name:     name,
			IsPreset: false,
		})
	}

	return categories, nil
}

// AddCategory 添加自定义分类
func (m *Manager) AddCategory(name string) error {
	// 检查是否为空
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("分类名称不能为空")
	}

	// 检查是否与预设分类重复
	for _, preset := range PresetCategories {
		if preset == name {
			return fmt.Errorf("不能添加与预设分类同名的分类")
		}
	}

	// 加载设置
	settings, err := m.storage.LoadSettings()
	if err != nil {
		return err
	}

	// 初始化自定义分类列表
	if settings.CustomCategories == nil {
		settings.CustomCategories = []string{}
	}

	// 检查是否已存在
	for _, existing := range settings.CustomCategories {
		if existing == name {
			return fmt.Errorf("分类已存在")
		}
	}

	// 添加分类
	settings.CustomCategories = append(settings.CustomCategories, name)

	// 保存设置
	return m.storage.SaveSettings(settings)
}

// DeleteCategory 删除自定义分类
// 返回使用该分类的 Skill ID 列表
func (m *Manager) DeleteCategory(name string) ([]string, error) {
	// 检查是否为预设分类
	for _, preset := range PresetCategories {
		if preset == name {
			return nil, fmt.Errorf("不能删除预设分类")
		}
	}

	// 加载设置
	settings, err := m.storage.LoadSettings()
	if err != nil {
		return nil, err
	}

	// 查找并删除分类
	found := false
	var newCategories []string
	for _, existing := range settings.CustomCategories {
		if existing == name {
			found = true
			continue
		}
		newCategories = append(newCategories, existing)
	}

	if !found {
		return nil, fmt.Errorf("分类不存在")
	}

	// 查找使用该分类的 Skills
	var skillsUsingCategory []string
	skills, _ := m.ListSkills()
	for _, skill := range skills {
		if skill.Category == name {
			skillsUsingCategory = append(skillsUsingCategory, skill.ID)
		}
	}

	// 将使用该分类的 Skills 改为"其他"
	for _, skillID := range skillsUsingCategory {
		meta, err := m.storage.LoadMetadata(skillID)
		if err != nil {
			continue
		}
		meta.Category = "其他"
		m.storage.SaveMetadata(skillID, meta)
	}

	// 更新设置
	settings.CustomCategories = newCategories
	if err := m.storage.SaveSettings(settings); err != nil {
		return nil, err
	}

	return skillsUsingCategory, nil
}
