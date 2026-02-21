package backend

import (
	"archive/zip"
	"context"
	"encoding/json"
	stdruntime "runtime"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"skillhub/backend/skill"
	"skillhub/backend/utils"
)

// App 主应用结构
type App struct {
	ctx      context.Context
	basePath string           // 添加 basePath 字段以跟踪当前存储路径
	manager  *skill.Manager
}

// 辅助函数：复制文件
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// 辅助函数：复制目录（递归复制，包括子目录）
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		// 如果是目录，创建对应目录
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// 复制文件
		srcData, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, srcData, info.Mode())
	})
}

// 辅助函数：创建 ZIP 文件
func createZip(files []string, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, file := range files {
		// 读取文件数据
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		// 在 ZIP 中创建文件条目
		writer, err := zipWriter.Create(filepath.Base(file))
		if err != nil {
			return err
		}

		// 写入数据到 ZIP 条目
		_, err = writer.Write(data)
		if err != nil {
			return err
		}
	}

	return nil
}

// 辅助函数：解压 ZIP 文件
func unzipFile(zipPath, destDir string) error {
	// 打开 ZIP 文件
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	// 创建目标目录
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// 解压文件
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(destDir, f.Name)

		// 创建目标文件（使用 os.Create 而不是 os.Open）
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// 写入解压数据
		_, err = io.Copy(file, rc)
		if err != nil {
			return err
		}
	}

	return nil
}

// NewApp 创建新的应用实例
func NewApp() *App {
	return &App{}
}

// Startup 在应用启动时调用
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// 获取默认 SkillHub 数据路径
	defaultPath := GetDefaultSkillHubPath()

	// 获取应用配置路径（%APPDATA%\SkillHub 或 ~/.config/skillhub）
	configPath := skill.GetAppConfigPath()

	// 1. 先尝试从配置路径读取 settings.json
	configStorage := skill.NewStorage(configPath)
	configSettings, err := configStorage.LoadSettings()

	var storagePath string

	if err == nil && configSettings.SkillHubPath != "" && !configSettings.FirstRun {
		// 找到已保存的配置，使用配置中的路径
		storagePath = configSettings.SkillHubPath
	} else {
		// 没有配置或首次运行，使用默认路径
		storagePath = defaultPath
	}

	// 2. 用确定的路径初始化管理器
	a.manager = skill.NewManager(storagePath)
	if err := a.manager.Initialize(); err != nil {
		// 如果初始化失败，回退到默认路径
		storagePath = defaultPath
		a.manager = skill.NewManager(storagePath)
		a.manager.Initialize()
	}

	// 3. 设置 basePath
	a.basePath = utils.ExpandPath(storagePath)

	// 4. 清理旧日志
	a.manager.GetStorage().CleanOldLogs()
}

// GetDefaultSkillHubPath 获取默认 SkillHub 路径
func GetDefaultSkillHubPath() string {
	home := utils.GetHomeDir()
	if stdruntime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
		if home == "" {
			home = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		}
		return filepath.Join(home, ".skill-hub")
	}
	return filepath.Join(os.Getenv("HOME"), ".skill-hub")
}

// === 前端绑定方法 ===

// GetSkills 获取所有 Skills
func (a *App) GetSkills() ([]skill.Skill, error) {
	return a.manager.ListSkills()
}

// GetSkill 获取单个 Skill
func (a *App) GetSkill(name string) (*skill.Skill, error) {
	return a.manager.GetSkill(name)
}

// ToggleSkill 切换 Skill 启用状态
func (a *App) ToggleSkill(skillName string, toolID string) (bool, error) {
	return a.manager.ToggleSkill(skillName, toolID)
}

// DeleteSkill 删除 Skill
func (a *App) DeleteSkill(name string) error {
	return a.manager.DeleteSkill(name)
}

// UpdateSkillMetadata 更新 Skill 元数据
func (a *App) UpdateSkillMetadata(skillID string, category string, tags []string, notes string) error {
	return a.manager.UpdateMetadata(skillID, category, tags, notes)
}

// InstallFromPath 从路径安装 Skill
func (a *App) InstallFromPath(path string, options skill.InstallOptions) (*skill.Skill, error) {
	return a.manager.InstallFromPath(path, options)
}

// ScanSkillPath 扫描 Skill 路径
func (a *App) ScanSkillPath(path string) (*skill.SkillInfo, error) {
	return a.manager.ScanSkillPath(path)
}

// ScanLocalPath 扫描本地路径（文件夹或压缩包）
func (a *App) ScanLocalPath(path string) (*skill.LocalScanResult, error) {
	return a.manager.ScanLocalPath(path)
}

// InstallFromLocalTemp 从本地临时目录安装 Skill
func (a *App) InstallFromLocalTemp(tempPath string, skillPath string, sourcePath string, options skill.InstallOptions) (*skill.Skill, error) {
	return a.manager.InstallFromLocalTemp(tempPath, skillPath, sourcePath, options)
}

// CleanupLocalTemp 清理本地临时目录
func (a *App) CleanupLocalTemp(tempPath string) {
	a.manager.CleanupLocalTemp(tempPath)
}

// === 工具相关方法 ===

// GetDetectedTools 获取检测到的工具
func (a *App) GetDetectedTools() []skill.ToolInfo {
	return a.manager.GetDetectedTools()
}

// GetToolAdapters 获取所有工具适配器信息
func (a *App) GetToolAdapters() []map[string]interface{} {
	var result []map[string]interface{}
	for _, adapter := range a.manager.GetRegistry().GetAll() {
		result = append(result, map[string]interface{}{
			"id":           adapter.ID(),
			"name":         adapter.Name(),
			"skills_path": adapter.SkillsPath(),
			"is_installed": adapter.IsInstalled(),
		})
	}
	return result
}

// === 配置管理方法 ===

// SetSkillHubPath 设置 SkillHub 路径
func (a *App) SetSkillHubPath(path string) error {
	a.manager = skill.NewManager(path)
	a.basePath = path // 同步更新 basePath，确保 GetStorageInfo 使用正确路径

	// 持久化到 settings.json，确保重启后保持新路径
	if err := a.manager.Initialize(); err != nil {
		return err
	}

	settings, err := a.GetSettings()
	if err != nil {
		return err
	}
	settings.SkillHubPath = path
	return a.SaveSettings(settings)
}

// GetSkillHubPath 获取当前 SkillHub 路径
func (a *App) GetSkillHubPath() string {
	return a.basePath  // 使用 a.basePath 而不是重新读取
}

// === 数据迁移相关方法 ===

// PathMigrationInfo 路径迁移信息
type PathMigrationInfo struct {
	HasOldData      bool   `json:"has_old_data"`       // 原路径是否有数据
	SkillsCount     int    `json:"skills_count"`         // 原路径 Skill 数量
	TotalSizeMB     int    `json:"total_size_mb"`       // 原路径总大小（MB）
	MigrationSizeMB int    `json:"migration_size_mb"`   // 需要迁移的数据大小
	OldPath         string `json:"old_path"`            // 原路径
	NewPath         string `json:"new_path"`            // 新路径
}

// GetMigrationInfo 获取路径迁移信息（用于弹窗提示）
func (a *App) GetMigrationInfo(newPath string) (*PathMigrationInfo, error) {
	oldPath := a.basePath

	// 如果新旧路径相同，无需迁移
	if oldPath == newPath {
		return &PathMigrationInfo{
			HasOldData: false,
			OldPath:     oldPath,
			NewPath:     newPath,
		}, nil
	}

	info := &PathMigrationInfo{
		OldPath: oldPath,
		NewPath: newPath,
	}

	oldExpanded := utils.ExpandPath(oldPath)

	// 添加：空路径保护
	if oldExpanded == "" {
		return nil, fmt.Errorf("原路径为空，无法获取迁移信息")
	}

	skillsCount := 0
	hasData := false

	// 1. 统计 Skills 数量（Skills 以目录形式存储在 skills/ 下）
	skillsDir := filepath.Join(oldExpanded, skill.SkillsDirName)
	if entries, err := os.ReadDir(skillsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				skillsCount++
			}
		}
		if skillsCount > 0 {
			hasData = true
		}
	}

	// 2. 检查其他关键数据是否存在（不计算大小，提升性能）
	otherPaths := []string{
		filepath.Join(oldExpanded, skill.MetadataDirName),
		filepath.Join(oldExpanded, skill.GitDirName),
		filepath.Join(oldExpanded, skill.HistoryDirName),
		filepath.Join(oldExpanded, skill.SettingsFile),
	}

	for _, path := range otherPaths {
		if hasDataAtPath(path) {
			hasData = true
		}
	}

	// 3. 检查 config/ 和 custom-tools.json
	configDir := filepath.Join(oldExpanded, "config")
	if hasDataAtPath(configDir) {
		hasData = true
	}

	customToolsPath := filepath.Join(oldExpanded, "custom-tools.json")
	if hasDataAtPath(customToolsPath) {
		hasData = true
	}

	// 4. 返回结果（简化：不计算大小，避免遍历整个目录树）
	info.SkillsCount = skillsCount
	info.TotalSizeMB = 0      // 简化：不计算大小
	info.MigrationSizeMB = 0
	info.HasOldData = hasData

	return info, nil
}

// hasDataAtPath 检查指定路径是否存在且有数据
func hasDataAtPath(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}

	// 如果是文件，存在即有数据
	if !stat.IsDir() {
		return true
	}

	// 如果是目录，检查是否为空
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	return len(entries) > 0
}

// getDirOrFileSize 获取目录或文件的大小（字节）
func getDirOrFileSize(path string) int64 {
	stat, err := os.Stat(path)
	if err != nil {
		return 0
	}

	// 如果是文件，直接返回大小
	if !stat.IsDir() {
		return stat.Size()
	}

	// 如果是目录，遍历计算总大小
	var totalSize int64
	filepath.Walk(path, func(subPath string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		totalSize += info.Size()
		return nil
	})

	return totalSize
}

// MigrateData 执行数据迁移（完整迁移所有数据）
func (a *App) MigrateData(newPath string) error {
	oldPath := a.basePath
	oldExpanded := utils.ExpandPath(oldPath)
	newExpanded := utils.ExpandPath(newPath)

	// 确保新路径目录存在
	if err := os.MkdirAll(newExpanded, 0755); err != nil {
		return fmt.Errorf("创建新路径失败: %w", err)
	}

	// 定义需要迁移的目录和文件
	dirsToMigrate := []string{
		"skills",     // Skills 本体
		"metadata",   // 元数据
		"git",        // Git 克隆缓存
		"history",    // 操作日志
		"config",     // 工具配置
	}

	// 迁移所有目录
	for _, dirName := range dirsToMigrate {
		oldDir := filepath.Join(oldExpanded, dirName)
		newDir := filepath.Join(newExpanded, dirName)
		if _, err := os.Stat(oldDir); err == nil {
			if err := copyDirectory(oldDir, newDir); err != nil {
				return fmt.Errorf("迁移 %s 失败: %w", dirName, err)
			}
		}
	}

	// 迁移 settings.json（保留原设置）
	oldSettings := filepath.Join(oldExpanded, "settings.json")
	newSettings := filepath.Join(newExpanded, "settings.json")
	if _, err := os.Stat(oldSettings); err == nil {
		data, err := os.ReadFile(oldSettings)
		if err != nil {
			return fmt.Errorf("读取 settings 失败: %w", err)
		}
		if err := os.WriteFile(newSettings, data, 0644); err != nil {
			return fmt.Errorf("复制 settings 失败: %w", err)
		}
	}

	// 迁移 custom-tools.json（自定义工具配置）
	oldCustomTools := filepath.Join(oldExpanded, "custom-tools.json")
	newCustomTools := filepath.Join(newExpanded, "custom-tools.json")
	if _, err := os.Stat(oldCustomTools); err == nil {
		data, err := os.ReadFile(oldCustomTools)
		if err != nil {
			return fmt.Errorf("读取 custom-tools 失败: %w", err)
		}
		if err := os.WriteFile(newCustomTools, data, 0644); err != nil {
			return fmt.Errorf("复制 custom-tools 失败: %w", err)
		}
	}

	return nil
}

// DeleteOldPath 删除原路径数据（仅删除内容，保留空目录）
func (a *App) DeleteOldPath(oldPath string) error {
	oldExpanded := utils.ExpandPath(oldPath)

	// 安全检查：避免删除系统目录
	defaultPath := GetDefaultSkillHubPath()
	if oldExpanded == "/" || oldExpanded == utils.ExpandPath(defaultPath) {
		// 如果是默认路径，询问用户确认（前端已处理，这里作为双重保险）
		return fmt.Errorf("为安全起见，不允许删除默认路径")
	}

	// 删除子目录内容
	dirs := []string{"skills", "metadata", "history", "git"}
	for _, dir := range dirs {
		dirPath := filepath.Join(oldExpanded, dir)
		if _, err := os.Stat(dirPath); err == nil {
			if err := os.RemoveAll(dirPath); err != nil {
				return fmt.Errorf("删除 %s 失败: %w", dir, err)
			}
		}
	}

	// 保留 settings.json 和空目录
	return nil
}

// SetSkillHubPathWithMigration 设置路径并迁移数据（原子操作）
func (a *App) SetSkillHubPathWithMigration(path string) error {
	// 1. 先迁移数据
	if err := a.MigrateData(path); err != nil {
		return err
	}

	// 2. 再切换路径
	return a.SetSkillHubPath(path)
}

// GetSettings 获取应用设置
func (a *App) GetSettings() (*skill.AppSettings, error) {
	return a.manager.GetStorage().LoadSettings()
}

// GetUpdateFrequency 获取更新频率设置
func (a *App) GetUpdateFrequency() string {
	settings, err := a.GetSettings()
	if err != nil {
		return "daily" // 默认每天检查
	}
	if settings.UpdateFrequency == "" {
		return "daily"
	}
	return settings.UpdateFrequency
}

// SetUpdateFrequency 设置更新频率
func (a *App) SetUpdateFrequency(frequency string) error {
	settings, err := a.GetSettings()
	if err != nil {
		return err
	}
	settings.UpdateFrequency = frequency
	return a.SaveSettings(settings)
}

// SaveSettings 保存应用设置
func (a *App) SaveSettings(settings *skill.AppSettings) error {
	return a.manager.GetStorage().SaveSettings(settings)
}

// GetOperationLogs 获取操作日志
func (a *App) GetOperationLogs() ([]skill.OperationLog, error) {
	return a.manager.GetStorage().ReadOperationLogs(time.Time{})
}

// ClearOperationLogs 清除操作日志
func (a *App) ClearOperationLogs() error {
	// 删除日志文件
	path := filepath.Join(a.manager.GetStorage().HistoryPath(), "operations.log")
	return os.Remove(path)
}

// InitializeWizard 初始化向导
func (a *App) InitializeWizard(storagePath string, selectedTools []string, customTools []*skill.CustomTool) error {
	err := a.manager.InitializeWizard(storagePath, selectedTools, customTools)
	if err != nil {
		return err
	}
	// 同步更新 basePath，确保后续调用 GetSkillHubPath 返回正确路径
	a.basePath = utils.ExpandPath(storagePath)
	return nil
}

// StorageInfo 存储空间信息
type StorageInfo struct {
	TotalSpace  int    `json:"total_space"`  // MB
	UsedSpace   int    `json:"used_space"`   // MB
	FreeSpace   int    `json:"free_space"`   // MB
	SkillsCount int    `json:"skills_count"` // Skill 数量
	SkillsPath  string `json:"skills_path"`  // Skills 目录路径
}

// GetStorageInfo 获取存储空间信息
func (a *App) GetStorageInfo() (*StorageInfo, error) {
	// 构建完整的存储路径
	expandedPath := utils.ExpandPath(a.basePath)

	// 统计目录信息
	var totalSize int64
	skillsCount := 0

	skillsDir := filepath.Join(expandedPath, "skills")
	if _, err := os.Stat(skillsDir); !os.IsNotExist(err) {
		// 先统计 Skills 数量（Skills 是目录）
		entries, err := os.ReadDir(skillsDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					skillsCount++
				}
			}
		}
		// 遍历 skills 目录统计大小
		err = filepath.Walk(skillsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// 获取磁盘分区的真实空间信息
	totalBytes, freeBytes, err := getDiskFreeSpace(expandedPath)
	if err != nil {
		// 如果获取失败，使用默认值作为降级处理
		const defaultTotalSpace = 100 * 1024 * 1024 * 1024 // 100GB
		const defaultFreeSpace = 50 * 1024 * 1024 * 1024   // 50GB
		totalBytes = defaultTotalSpace
		freeBytes = defaultFreeSpace
	}
	usedBytes := totalBytes - freeBytes

	return &StorageInfo{
		TotalSpace:  convertBytesToMB(int64(totalBytes)),
		UsedSpace:   convertBytesToMB(int64(usedBytes)),
		FreeSpace:   convertBytesToMB(int64(freeBytes)),
		SkillsCount: skillsCount,
		SkillsPath:  skillsDir,
	}, nil
}

// convertBytesToMB 将字节转换为MB
func convertBytesToMB(bytes int64) int {
	return int(bytes / (1024 * 1024))
}

// GetCustomTools 获取自定义工具列表
func (a *App) GetCustomTools() ([]*skill.CustomTool, error) {
	return a.manager.GetCustomTools()
}

// RemoveCustomTool 删除自定义工具
func (a *App) RemoveCustomTool(id string) error {
	return a.manager.RemoveCustomTool(id)
}

// ToggleToolEnabled 切换工具启用状态
func (a *App) ToggleToolEnabled(toolID string) error {
	return a.manager.ToggleToolEnabled(toolID)
}

// AddCustomTool 添加自定义工具
func (a *App) AddCustomTool(tool *skill.CustomTool) error {
	return a.manager.AddCustomTool(tool)
}

// === Git 安装相关方法 ===

// GitURLInfo Git URL 信息
type GitURLInfo struct {
	Owner    string `json:"owner"`
	Repo     string `json:"repo"`
	SubPath  string `json:"sub_path"`
	FullURL  string `json:"full_url"`
	ShortRef string `json:"short_ref"`
}

// ParseGitURL 解析 Git URL
func (a *App) ParseGitURL(url string) (*GitURLInfo, error) {
	info, err := a.manager.ParseGitURL(url)
	if err != nil {
		return nil, err
	}
	return &GitURLInfo{
		Owner:    info.Owner,
		Repo:     info.Repo,
		SubPath:  info.SubPath,
		FullURL:  info.FullURL,
		ShortRef: info.ShortRef,
	}, nil
}

// CloneFromGit 从 Git 克隆并扫描 Skills
func (a *App) CloneFromGit(url string) (*skill.GitInstallResult, error) {
	return a.manager.CloneFromGit(url)
}

// InstallFromGit 从 Git 临时目录安装 Skill
func (a *App) InstallFromGit(tempPath string, skillPath string, gitURL string, options skill.InstallOptions) (*skill.Skill, error) {
	return a.manager.InstallFromGit(tempPath, skillPath, gitURL, options)
}

// CleanupClone 清理克隆的临时目录
func (a *App) CleanupClone(tempPath string) {
	a.manager.CleanupClone(tempPath)
}

// FetchSkillsLeaderboard 获取 skills.sh 热门榜单
type SkillsLeaderboardItem struct {
	Rank    int    `json:"rank"`
	Name    string `json:"name"`
	Author  string `json:"author"`
	Installs string `json:"installs"`
	URL     string `json:"url"`
}

func (a *App) FetchSkillsLeaderboard(listType string) ([]SkillsLeaderboardItem, error) {
	client := utils.NewSkillsClient()

	// 实时从 skills.sh 抓取数据
	leaderboard, err := client.FetchLeaderboard(listType)
	if err != nil {
		return nil, err
	}

	// 转换返回类型
	var result []SkillsLeaderboardItem
	for _, item := range leaderboard.Items {
		result = append(result, SkillsLeaderboardItem{
			Rank:    item.Rank,
			Name:    item.Name,
			Author:  item.Author,
			Installs: item.Installs,
			URL:     item.URL,
		})

	}

	return result, nil
}

// === 导出/导入相关方法 ===

// ExportInfo 导出信息
type ExportInfo struct {
	SkillsCount      int    `json:"skills_count"`
	GitCacheCount    int    `json:"git_cache_count"`
	CustomToolsCount int    `json:"custom_tools_count"`
	EstimatedSize    string `json:"estimated_size"`
}

// GetExportInfo 获取导出信息
func (a *App) GetExportInfo() (*ExportInfo, error) {
	expandedPath := utils.ExpandPath(a.basePath)
	info := &ExportInfo{}

	// 统计 Skills 数量
	skillsDir := filepath.Join(expandedPath, "skills")
	if entries, err := os.ReadDir(skillsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				info.SkillsCount++
			}
		}
	}

	// 统计 Git 缓存数量
	gitDir := filepath.Join(expandedPath, "git")
	if entries, err := os.ReadDir(gitDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				info.GitCacheCount++
			}
		}
	}

	// 统计自定义工具数量
	customTools, err := a.manager.GetCustomTools()
	if err == nil {
		info.CustomToolsCount = len(customTools)
	}

	// 计算预计大小
	var totalSize int64
	dirsToCheck := []string{"skills", "git", "config", "metadata"}
	for _, dirName := range dirsToCheck {
		dirPath := filepath.Join(expandedPath, dirName)
		filepath.Walk(dirPath, func(path string, f os.FileInfo, err error) error {
			if err != nil || f.IsDir() {
				return err
			}
			totalSize += f.Size()
			return nil
		})
	}

	// 转换为人类可读格式
	if totalSize < 1024*1024 {
		info.EstimatedSize = fmt.Sprintf("~%d KB", totalSize/1024)
	} else if totalSize < 1024*1024*1024 {
		info.EstimatedSize = fmt.Sprintf("~%d MB", totalSize/(1024*1024))
	} else {
		info.EstimatedSize = fmt.Sprintf("~%.1f GB", float64(totalSize)/(1024*1024*1024))
	}

	return info, nil
}

// ImportPreview 导入预览信息
type ImportPreview struct {
	Version          string `json:"version"`
	ExportDate       string `json:"export_date"`
	SkillsCount      int    `json:"skills_count"`
	GitCacheCount    int    `json:"git_cache_count"`
	CustomToolsCount int    `json:"custom_tools_count"`
}

// PreviewImportFile 预览导入文件
func (a *App) PreviewImportFile(filePath string) (*ImportPreview, error) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "skillhub-preview-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// 解压 ZIP 文件
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开 ZIP 文件失败: %w", err)
	}
	defer r.Close()

	// 解压文件
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}

		path := filepath.Join(tempDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), 0755)
			file, err := os.Create(path)
			if err != nil {
				rc.Close()
				return nil, err
			}
			io.Copy(file, rc)
			file.Close()
		}
		rc.Close()
	}

	preview := &ImportPreview{
		Version: "1.0",
	}

	// 读取 manifest.json
	manifestPath := filepath.Join(tempDir, "manifest.json")
	if data, err := os.ReadFile(manifestPath); err == nil {
		var manifest struct {
			Version    string `json:"skillhub_version"`
			ExportDate string `json:"export_date"`
			Skills     []struct {
				Name string `json:"name"`
			} `json:"skills"`
			CustomTools []struct {
				Name string `json:"name"`
			} `json:"custom_tools"`
			GitCache []struct {
				RepoURL string `json:"repo_url"`
			} `json:"git_cache"`
		}
		if err := json.Unmarshal(data, &manifest); err == nil {
			preview.Version = manifest.Version
			preview.ExportDate = manifest.ExportDate
			preview.SkillsCount = len(manifest.Skills)
			preview.CustomToolsCount = len(manifest.CustomTools)
			preview.GitCacheCount = len(manifest.GitCache)
		}
	} else {
		// 如果没有 manifest.json，从目录结构推断
		// 统计 Skills 数量
		skillsDir := filepath.Join(tempDir, "skills")
		if entries, err := os.ReadDir(skillsDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					preview.SkillsCount++
				}
			}
		}

		// 统计 Git 缓存数量
		gitDir := filepath.Join(tempDir, "git")
		if entries, err := os.ReadDir(gitDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					preview.GitCacheCount++
				}
			}
		}

		// 使用当前日期
		preview.ExportDate = time.Now().Format(time.RFC3339)
	}

	return preview, nil
}

// ExportData 导出数据到 ZIP 文件
func (a *App) ExportData(exportSkills bool, exportSettings bool) (string, error) {
	expandedPath := utils.ExpandPath(a.basePath)

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "skillhub-export-*")
	if err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 创建导出目录
	exportDir := filepath.Join(tempDir, "skillhub-export")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("创建导出目录失败: %w", err)
	}

	// 1. 导出 Skills
	if exportSkills {
		skillsSrc := filepath.Join(expandedPath, "skills")
		if _, err := os.Stat(skillsSrc); err == nil {
			destDir := filepath.Join(exportDir, "skills")
			if err := copyDirectory(skillsSrc, destDir); err != nil {
				os.RemoveAll(tempDir)
				return "", fmt.Errorf("导出 Skills 失败: %w", err)
			}
		}
	}

	// 2. 导出 Git 缓存
	gitSrc := filepath.Join(expandedPath, "git")
	if _, err := os.Stat(gitSrc); err == nil {
		destDir := filepath.Join(exportDir, "git")
		if err := copyDirectory(gitSrc, destDir); err != nil {
			os.RemoveAll(tempDir)
			return "", fmt.Errorf("导出 Git 缓存失败: %w", err)
		}
	}

	// 3. 导出配置目录
	configSrc := filepath.Join(expandedPath, "config")
	if _, err := os.Stat(configSrc); err == nil {
		destDir := filepath.Join(exportDir, "config")
		if err := copyDirectory(configSrc, destDir); err != nil {
			os.RemoveAll(tempDir)
			return "", fmt.Errorf("导出配置失败: %w", err)
		}
	}

	// 4. 导出设置
	if exportSettings {
		settingsSrc := filepath.Join(expandedPath, "settings.json")
		if _, err := os.Stat(settingsSrc); err == nil {
			destFile := filepath.Join(exportDir, "settings.json")
			if err := copyFile(settingsSrc, destFile); err != nil {
				os.RemoveAll(tempDir)
				return "", fmt.Errorf("导出设置失败: %w", err)
			}
		}

		// 导出自定义工具配置
		customToolsSrc := filepath.Join(expandedPath, "custom-tools.json")
		if _, err := os.Stat(customToolsSrc); err == nil {
			destFile := filepath.Join(exportDir, "custom-tools.json")
			if err := copyFile(customToolsSrc, destFile); err != nil {
				os.RemoveAll(tempDir)
				return "", fmt.Errorf("导出自定义工具配置失败: %w", err)
			}
		}

		// 导出元数据
		metadataSrc := filepath.Join(expandedPath, "metadata")
		if _, err := os.Stat(metadataSrc); err == nil {
			destDir := filepath.Join(exportDir, "metadata")
			if err := copyDirectory(metadataSrc, destDir); err != nil {
				os.RemoveAll(tempDir)
				return "", fmt.Errorf("导出元数据失败: %w", err)
			}
		}
	}

	// 5. 创建 manifest.json
	manifest := struct {
		Version     string `json:"skillhub_version"`
		ExportDate  string `json:"export_date"`
		Skills      []map[string]interface{} `json:"skills"`
		CustomTools []map[string]interface{} `json:"custom_tools"`
		GitCache    []map[string]interface{} `json:"git_cache"`
		Settings    map[string]interface{} `json:"settings"`
	}{
		Version:    "1.0",
		ExportDate: time.Now().Format(time.RFC3339),
		Skills:     []map[string]interface{}{},
		CustomTools: []map[string]interface{}{},
		GitCache:   []map[string]interface{}{},
		Settings:   map[string]interface{}{},
	}

	// 读取 Skills 元数据
	skillNames, _ := a.manager.GetStorage().ListSkills()
	for _, name := range skillNames {
		meta, err := a.manager.GetStorage().LoadMetadata(name)
		if err == nil {
			skillInfo := map[string]interface{}{
				"name":            name,
				"source":          meta.SourceURL,
				"source_type":     meta.SourceType,
				"tags":            meta.Tags,
				"notes":           meta.Notes,
				"category":        meta.Category,
				"installed_at":    meta.InstalledAt.Format(time.RFC3339),
				"original_name":   meta.OriginalName,
				"author":          meta.Author,
			}
			manifest.Skills = append(manifest.Skills, skillInfo)
		}
	}

	// 读取自定义工具
	customTools, _ := a.manager.GetCustomTools()
	for _, tool := range customTools {
		toolInfo := map[string]interface{}{
			"name":       tool.Name,
			"path":       tool.SkillsPath,
			"added_date": tool.DateAdded,
		}
		manifest.CustomTools = append(manifest.CustomTools, toolInfo)
	}

	// 统计 Git 缓存
	gitDir := filepath.Join(expandedPath, "git")
	if entries, err := os.ReadDir(gitDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				gitInfo := map[string]interface{}{
					"local_path": filepath.Join("git", entry.Name()),
				}
				manifest.GitCache = append(manifest.GitCache, gitInfo)
			}
		}
	}

	// 读取设置
	settings, _ := a.GetSettings()
	if settings != nil {
		manifest.Settings = map[string]interface{}{
			"skillhub_path":     settings.SkillHubPath,
			"theme":             settings.Theme,
			"auto_update_check": settings.AutoUpdateCheck,
			"update_frequency":  settings.UpdateFrequency,
		}
	}

	// 写入 manifest.json
	manifestPath := filepath.Join(exportDir, "manifest.json")
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("写入 manifest 失败: %w", err)
	}

	// 6. 创建 ZIP 文件
	homeDir := utils.GetHomeDir()
	downloadsDir := filepath.Join(homeDir, "Downloads")
	if stdruntime.GOOS == "windows" {
		downloadsDir = filepath.Join(os.Getenv("USERPROFILE"), "Downloads")
	}

	// 确保 Downloads 目录存在
	os.MkdirAll(downloadsDir, 0755)

	zipFileName := fmt.Sprintf("skillhub-backup-%s.zip", time.Now().Format("2006-01-02"))
	zipPath := filepath.Join(downloadsDir, zipFileName)

	// 创建 ZIP 文件
	zipFile, err := os.Create(zipPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("创建 ZIP 文件失败: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 遍历导出目录，添加所有文件到 ZIP
	err = filepath.Walk(exportDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(exportDir, path)
		if err != nil {
			return err
		}

		// 创建 ZIP 条目
		if info.IsDir() {
			_, err = zipWriter.Create(relPath + "/")
			return err
		}

		// 添加文件
		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("创建 ZIP 失败: %w", err)
	}

	// 清理临时目录
	os.RemoveAll(tempDir)

	return zipPath, nil
}

// ImportData 从 ZIP 文件导入数据
func (a *App) ImportData(filePath string, mergeSkills bool, mergeSettings bool) error {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "skillhub-import-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// 解压 ZIP
	if err := unzipFile(filePath, tempDir); err != nil {
		return fmt.Errorf("解压文件失败: %w", err)
	}

	// 读取导入的设置
	var importSettings skill.AppSettings
	settingsData, err := os.ReadFile(filepath.Join(tempDir, "settings.json"))
	if err == nil {
		if err := json.Unmarshal(settingsData, &importSettings); err != nil {
			return err
		}
	}

	// 读取导入的 Skills 元数据
	var importedSkills map[string]skill.Metadata
	skillsData, err := os.ReadFile(filepath.Join(tempDir, "skills", "skills.json"))
	if err == nil {
		if err := json.Unmarshal(skillsData, &importedSkills); err != nil {
			return err
		}
	}

	// 获取当前 Skills 名称列表
	currentSkillNames, err := a.manager.GetStorage().ListSkills()
	if err != nil {
		return err
	}

	// 构建当前 Skills 的元数据映射
	currentSkillsMap := make(map[string]skill.Metadata)
	for _, name := range currentSkillNames {
		meta, err := a.manager.GetStorage().LoadMetadata(name)
		if err == nil && meta != nil {
			currentSkillsMap[name] = *meta // 解引用指针
		}
	}

	// 合并或覆盖
	var finalSkills map[string]skill.Metadata
	if mergeSkills {
		// 覆盖模式：使用导入的 Skills
		finalSkills = importedSkills
	} else {
		// 合并模式：保留现有 Skills，添加导入的
		finalSkills = make(map[string]skill.Metadata)
		for name, meta := range currentSkillsMap {
			finalSkills[name] = meta
		}
		for name, meta := range importedSkills {
			finalSkills[name] = meta
		}
	}

	// 保存 Skills 元数据
	for name, meta := range finalSkills {
		metaPath := a.manager.GetStorage().SkillMetadataPath(name)

		// 读取现有元数据进行合并
		var existingMeta skill.Metadata
		if metaBytes, err := os.ReadFile(metaPath); err == nil {
			json.Unmarshal(metaBytes, &existingMeta)
		}

		// 合并元数据：现有值优先
		if existingMeta.Category != "" {
			meta.Category = existingMeta.Category
		}
		if existingMeta.Tags != nil && len(existingMeta.Tags) > 0 {
			meta.Tags = existingMeta.Tags
		}
		if existingMeta.Notes != "" {
			meta.Notes = existingMeta.Notes
		}

		// 保存
		metaBytes, _ := json.MarshalIndent(meta, "", "  ")
		if err := os.WriteFile(metaPath, metaBytes, 0644); err != nil {
			return fmt.Errorf("保存元数据失败: %w", err)
		}
	}

	// 保存导入的设置
	if mergeSettings {
		// 使用 skill.AppSettings 类型
		settings := &skill.AppSettings{
			FirstRun:        false, // 导入后标记为非首次
			SkillHubPath:    importSettings.SkillHubPath,
			Theme:           importSettings.Theme,
			AutoUpdateCheck: importSettings.AutoUpdateCheck,
			UpdateFrequency: importSettings.UpdateFrequency,
		}
		if err := a.manager.GetStorage().SaveSettings(settings); err != nil {
			return fmt.Errorf("保存设置失败: %w", err)
		}
	}

	return nil
}

// === SKILL.md 相关方法 ===

// ReadSkillFile 读取 Skill 文件内容
func (a *App) ReadSkillFile(skillName string, filename string) (string, error) {
	path := filepath.Join(a.manager.GetStorage().SkillPath(skillName), filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取 Skill 文件失败: %w", err)
	}

	return string(data), nil
}

// ListSkillFiles 列出 Skill 目录中的所有文件
type FileInfo struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	IsDir    bool      `json:"is_dir"`
	Size      int64      `json:"size"`
	Modified  time.Time  `json:"modified"`
}

func (a *App) ListSkillFiles(skillName string, subPath string) ([]FileInfo, error) {
	skillPath := a.manager.GetStorage().SkillPath(skillName)

	// 构建目标路径
	var targetPath string
	if subPath != "" {
		targetPath = filepath.Join(skillPath, subPath)
	} else {
		targetPath = skillPath
	}

	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, FileInfo{
			Name:     entry.Name(),
			Path:     filepath.Join(targetPath, entry.Name()),
			IsDir:    entry.IsDir(),
			Size:     info.Size(),
			Modified: info.ModTime(),
		})
	}

	return files, nil
}

// === 浏览相关方法 ===

// SelectFolder 选择文件夹
func (a *App) SelectFolder() (string, error) {
	// 检查 context 是否已初始化
	if a.ctx == nil {
		return "", fmt.Errorf("应用上下文未初始化")
	}

	// 使用文件对话框（现在使用 64 位架构）
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 Skill 文件夹",
	})

	if err != nil {
		return "", fmt.Errorf("打开文件夹对话框失败: %w", err)
	}

	// 用户取消时返回空字符串
	return selection, nil
}

// SelectFile 选择文件
func (a *App) SelectFile() (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "选择备份文件",
		Filters: []runtime.FileFilter{{DisplayName: "ZIP 文件", Pattern: "*.zip"}},
	})
	if err != nil {
		return "", err
	}
	if selection == "" {
		return "", nil
	}

	return selection, nil
}

// SelectInstallFile 选择安装文件（支持 zip 和 skill 格式）
func (a *App) SelectInstallFile() (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 Skill 文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "Skill 文件", Pattern: "*.zip;*.skill"},
			{DisplayName: "ZIP 压缩包", Pattern: "*.zip"},
			{DisplayName: "Skill 包", Pattern: "*.skill"},
		},
	})
	if err != nil {
		return "", err
	}
	if selection == "" {
		return "", nil
	}

	return selection, nil
}

// === 路径验证相关方法 ===

// PathValidationResult 路径验证结果
type PathValidationResult struct {
	IsValid    bool    `json:"is_valid"`
	IsWritable bool    `json:"is_writable"`
	DiskFreeGB float64 `json:"disk_free_gb"`
	Message    string  `json:"message"`
	Status     string  `json:"status"` // "valid", "warning", "error"
}

// ValidateStoragePath 验证存储路径
func (a *App) ValidateStoragePath(path string) (*PathValidationResult, error) {
	result := &PathValidationResult{
		IsValid:    false,
		IsWritable: false,
		DiskFreeGB: 0,
		Status:     "error",
	}

	// 展开路径中的 ~
	expandedPath := utils.ExpandPath(path)

	// 0. 检查是否为绝对路径（格式验证）
	if !filepath.IsAbs(expandedPath) {
		result.Message = getAbsolutePathHint()
		return result, nil
	}

	// 1. 检查路径是否存在，不存在则尝试创建
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		// 尝试创建目录
		if err := os.MkdirAll(expandedPath, 0755); err != nil {
			result.Message = fmt.Sprintf("无法创建目录: %v", err)
			return result, nil
		}
	}

	// 2. 检查写入权限
	testFile := filepath.Join(expandedPath, ".skillhub_write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		result.Message = "路径不可写，请检查权限"
		return result, nil
	}
	os.Remove(testFile) // 清理测试文件
	result.IsWritable = true

	// 3. 检查磁盘空间
	var freeSpace uint64
	var statPath string

	// 获取要检查的磁盘路径
	fileInfo, err := os.Stat(expandedPath)
	if err != nil {
		result.Message = fmt.Sprintf("无法访问路径: %v", err)
		return result, nil
	}

	if fileInfo.IsDir() {
		statPath = expandedPath
	} else {
		statPath = filepath.Dir(expandedPath)
	}

	// 跨平台获取磁盘空间
	_, freeSpace, err = getDiskFreeSpace(statPath)
	if err != nil {
		// 如果获取失败，假设空间充足（某些系统可能不支持）
		freeSpace = 10 * 1024 * 1024 * 1024 // 10GB
	}

	result.DiskFreeGB = float64(freeSpace) / (1024 * 1024 * 1024)

	// 4. 判断验证结果
	result.IsValid = true

	if result.DiskFreeGB < 0.5 {
		result.Status = "error"
		result.Message = fmt.Sprintf("磁盘空间不足 (%.1f GB)，建议至少 500 MB", result.DiskFreeGB)
	} else if result.DiskFreeGB < 2 {
		result.Status = "warning"
		result.Message = fmt.Sprintf("磁盘空间偏低 (%.1f GB 可用)", result.DiskFreeGB)
	} else {
		result.Status = "valid"
		result.Message = fmt.Sprintf("%.1f GB 可用", result.DiskFreeGB)
	}

	return result, nil
}

// getDiskFreeSpace 获取磁盘空间信息（跨平台分发）
// 返回：总空间（字节）、可用空间（字节）
func getDiskFreeSpace(path string) (totalBytes, freeBytes uint64, err error) {
	if stdruntime.GOOS == "windows" {
		return getDiskFreeSpaceWindows(path)
	}
	return getDiskFreeSpaceUnix(path)
}

// getAbsolutePathHint 返回绝对路径格式提示（跨平台）
func getAbsolutePathHint() string {
	if stdruntime.GOOS == "windows" {
		return "请输入绝对路径，如：C:\\skill-hub 或 D:\\data"
	}
	return "请输入绝对路径，如：/home/user/skill-hub 或 ~/skill-hub"
}

// === 数据管理相关方法 ===

// UpdateInfo 更新信息（用于应用和 Skills）
type UpdateInfo struct {
	// 应用更新字段（保留原有字段）
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
	DownloadURL    string `json:"download_url,omitempty"`
	ReleaseNotes   string `json:"release_notes,omitempty"`

	// Skill 更新字段
	SkillsWithUpdate []string `json:"skills_with_update"`
	UpdateCount      int      `json:"update_count"`
}

// ResetAllData 重置所有数据
func (a *App) ResetAllData() error {
	expandedPath := utils.ExpandPath(a.basePath)

	// 清空 skills 目录
	skillsDir := filepath.Join(expandedPath, "skills")
	if _, err := os.Stat(skillsDir); err == nil {
		entries, err := os.ReadDir(skillsDir)
		if err != nil {
			return fmt.Errorf("读取 skills 目录失败: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				skillPath := filepath.Join(skillsDir, entry.Name())
				if err := os.RemoveAll(skillPath); err != nil {
					return fmt.Errorf("删除 skill 失败: %w", err)
				}
			}
		}
	}

	// 清空元数据目录
	metadataDir := filepath.Join(expandedPath, "metadata")
	if _, err := os.Stat(metadataDir); err == nil {
		entries, err := os.ReadDir(metadataDir)
		if err != nil {
			return fmt.Errorf("读取 metadata 目录失败: %w", err)
		}
		for _, entry := range entries {
			metaPath := filepath.Join(metadataDir, entry.Name())
			if err := os.Remove(metaPath); err != nil {
				return fmt.Errorf("删除元数据失败: %w", err)
			}
		}
	}

	// 清空操作日志
	logPath := filepath.Join(expandedPath, "history", "operations.log")
	if _, err := os.Stat(logPath); err == nil {
		if err := os.Remove(logPath); err != nil {
			return fmt.Errorf("删除日志失败: %w", err)
		}
	}

	return nil
}

// CheckForUpdates 检查 Skill 更新
func (a *App) CheckForUpdates() (*UpdateInfo, error) {
	// 检查更新设置
	settings, err := a.GetSettings()
	if err != nil {
		return nil, err
	}

	// 如果未启用自动检查，返回空结果
	if !settings.AutoUpdateCheck {
		return &UpdateInfo{
			SkillsWithUpdate: []string{},
			UpdateCount:      0,
		}, nil
	}

	// 检查所有 Skills 的更新
	updates, err := a.manager.CheckSkillUpdates()
	if err != nil {
		return nil, err
	}

	// 构建结果
	var skillNames []string
	for _, u := range updates {
		if u.HasUpdate {
			skillNames = append(skillNames, u.Name)
		}
	}

	return &UpdateInfo{
		SkillsWithUpdate: skillNames,
		UpdateCount:      len(skillNames),
	}, nil
}

// UpdateSingleSkill 更新单个 Skill
func (a *App) UpdateSingleSkill(skillName string) error {
	return a.manager.UpdateSkill(skillName)
}

// === 分类管理相关方法 ===

// GetAllCategories 获取所有分类（预设 + 自定义）
func (a *App) GetAllCategories() ([]skill.CategoryInfo, error) {
	return a.manager.GetAllCategories()
}

// AddCategory 添加自定义分类
func (a *App) AddCategory(name string) error {
	return a.manager.AddCategory(name)
}

// DeleteCategory 删除自定义分类
// 返回使用该分类的 Skill ID 列表
func (a *App) DeleteCategory(name string) ([]string, error) {
	return a.manager.DeleteCategory(name)
}
