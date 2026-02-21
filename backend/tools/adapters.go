package tools

import (
	"os"
	"path/filepath"
	"runtime"
)

// ClaudeAdapter Claude Code 适配器
type ClaudeAdapter struct {
	*BaseAdapter
}

func NewClaudeAdapter() *ClaudeAdapter {
	return &ClaudeAdapter{
		BaseAdapter: NewBaseAdapter(
			"claude-code",
			"Claude Code",
			getClaudeSkillsPath(),
		),
	}
}

func getClaudeSkillsPath() string {
	home := getHomeDir()
	return filepath.Join(home, ".claude", "skills")
}

// OpenCodeAdapter OpenCode 适配器
type OpenCodeAdapter struct {
	*BaseAdapter
}

func NewOpenCodeAdapter() *OpenCodeAdapter {
	return &OpenCodeAdapter{
		BaseAdapter: NewBaseAdapter(
			"opencode",
			"OpenCode",
			getOpenCodeSkillsPath(),
		),
	}
}

func getOpenCodeSkillsPath() string {
	home := getHomeDir()
	return filepath.Join(home, ".config", "opencode", "skills")
}

// CursorAdapter Cursor 适配器
type CursorAdapter struct {
	*BaseAdapter
}

func NewCursorAdapter() *CursorAdapter {
	return &CursorAdapter{
		BaseAdapter: NewBaseAdapter(
			"cursor",
			"Cursor",
			getCursorSkillsPath(),
		),
	}
}

func getCursorSkillsPath() string {
	home := getHomeDir()
	return filepath.Join(home, ".cursor", "skills")
}

// CodeBuddyAdapter CodeBuddy 适配器
type CodeBuddyAdapter struct {
	*BaseAdapter
}

func NewCodeBuddyAdapter() *CodeBuddyAdapter {
	return &CodeBuddyAdapter{
		BaseAdapter: NewBaseAdapter(
			"codebuddy",
			"CodeBuddy",
			getCodeBuddySkillsPath(),
		),
	}
}

func getCodeBuddySkillsPath() string {
	home := getHomeDir()
	return filepath.Join(home, ".codebuddy", "skills")
}

// TraeAdapter Trae 适配器
type TraeAdapter struct {
	*BaseAdapter
}

func NewTraeAdapter() *TraeAdapter {
	return &TraeAdapter{
		BaseAdapter: NewBaseAdapter(
			"trae",
			"Trae",
			getTraeSkillsPath(),
		),
	}
}

func getTraeSkillsPath() string {
	home := getHomeDir()
	return filepath.Join(home, ".trae", "skills")
}

// getHomeDir 跨平台获取用户主目录
func getHomeDir() string {
	if runtime.GOOS == "windows" {
		// Windows: 使用 USERPROFILE
		home := os.Getenv("USERPROFILE")
		if home == "" {
			home = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		}
		return home
	}
	// Unix-like: 使用 HOME
	return os.Getenv("HOME")
}
