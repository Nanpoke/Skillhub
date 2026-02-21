package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitHubRelease GitHub Release API 响应
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// GitClient Git 操作客户端
type GitClient struct{}

// NewGitClient 创建 Git 客户端
func NewGitClient() *GitClient {
	return &GitClient{}
}

// IsGitInstalled 检查 Git 是否已安装
func (g *GitClient) IsGitInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// CloneResult 克隆结果
type CloneResult struct {
	TempPath string   // 临时目录路径
	Skills   []string // 发现的 Skill 名称列表
	Error    string   // 错误信息
}

// Clone 浅克隆仓库到临时目录
func (g *GitClient) Clone(url string) (*CloneResult, error) {
	if !g.IsGitInstalled() {
		return nil, fmt.Errorf("Git 未安装，请先安装 Git")
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "skillhub-clone-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 执行 git clone --depth 1
	cmd := exec.Command("git", "clone", "--depth", "1", url, tempDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("克隆失败: %s\n%s", err.Error(), string(output))
	}

	// 扫描 Skill 目录
	skills := g.ScanSkills(tempDir)

	return &CloneResult{
		TempPath: tempDir,
		Skills:   skills,
	}, nil
}

// ScanSkills 扫描目录中的 Skills
func (g *GitClient) ScanSkills(dir string) []string {
	var skills []string

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// 检查是否是 SKILL.md 文件
		if !info.IsDir() && info.Name() == "SKILL.md" {
			// 获取包含 SKILL.md 的目录路径
			skillDir := filepath.Dir(path)
			relPath, _ := filepath.Rel(dir, skillDir)

			if relPath == "." {
				// 根目录就是 Skill
				skills = append(skills, "")
			} else {
				skills = append(skills, relPath)
			}
		}

		return nil
	})

	return skills
}

// ParseGitURL 解析 Git URL
type GitURLInfo struct {
	Owner    string // 仓库所有者
	Repo     string // 仓库名
	SubPath  string // 子路径（可选）
	FullURL  string // 完整 URL
	ShortRef string // 简短引用 (owner/repo)
}

// ParseGitURL 解析各种格式的 Git URL
func (g *GitClient) ParseGitURL(input string) (*GitURLInfo, error) {
	input = strings.TrimSpace(input)

	info := &GitURLInfo{
		FullURL: input,
	}

	// 格式1: https://github.com/owner/repo/tree/branch/path
	if strings.HasPrefix(input, "https://github.com/") {
		// 移除 https://github.com/ 前缀
		rest := strings.TrimPrefix(input, "https://github.com/")

		// 检查是否有 tree/ 分支路径
		if strings.Contains(rest, "/tree/") {
			parts := strings.SplitN(rest, "/tree/", 2)
			ownerRepo := parts[0]
			info.ShortRef = ownerRepo

			ownerRepoParts := strings.SplitN(ownerRepo, "/", 2)
			if len(ownerRepoParts) == 2 {
				info.Owner = ownerRepoParts[0]
				info.Repo = ownerRepoParts[1]
			}

			// 提取子路径（跳过分支名）
			if len(parts) == 2 {
				treeParts := strings.SplitN(parts[1], "/", 2)
				if len(treeParts) == 2 {
					info.SubPath = treeParts[1]
				}
			}

			// 构建基础克隆 URL
			info.FullURL = fmt.Sprintf("https://github.com/%s", ownerRepo)
		} else {
			info.ShortRef = rest
			parts := strings.SplitN(rest, "/", 2)
			if len(parts) == 2 {
				info.Owner = parts[0]
				info.Repo = strings.TrimSuffix(parts[1], ".git")
			}
			info.FullURL = fmt.Sprintf("https://github.com/%s", info.ShortRef)
		}
		return info, nil
	}

	// 格式2: owner/repo (GitHub 简写)
	if strings.Contains(input, "/") && !strings.HasPrefix(input, "http") {
		parts := strings.SplitN(input, "/", 2)
		if len(parts) == 2 {
			info.Owner = parts[0]
			info.Repo = parts[1]
			info.ShortRef = input
			info.FullURL = fmt.Sprintf("https://github.com/%s", input)
		}
		return info, nil
	}

	return nil, fmt.Errorf("无法解析 Git URL: %s", input)
}

// Pull 拉取更新
func (g *GitClient) Pull(repoPath string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("拉取失败: %s", string(output))
	}
	return nil
}

// GetRemoteURL 获取远程仓库 URL
func (g *GitClient) GetRemoteURL(repoPath string) (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// MoveGitDir 移动 .git 目录
func (g *GitClient) MoveGitDir(srcPath, destPath string) error {
	srcGit := filepath.Join(srcPath, ".git")
	destGit := filepath.Join(destPath, ".git")

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(destGit), 0755); err != nil {
		return err
	}

	// 移动 .git 目录
	return os.Rename(srcGit, destGit)
}

// GetTag 获取当前目录的 Git tag
func (g *GitClient) GetTag(repoPath string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("获取 tag 失败: %s", string(output))
	}
	return strings.TrimSpace(string(output)), nil
}

// FetchLatestRelease 从 GitHub API 获取最新 release
func (g *GitClient) FetchLatestRelease(owner, repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API 返回错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	return &release, nil
}

// CompareVersions 比较两个版本号
// 返回: -1 (v1 < v2), 0 (v1 == v2), 1 (v1 > v2)
func (g *GitClient) CompareVersions(v1, v2 string) int {
	// 移除 "v" 前缀
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// 按点分割
	p1 := strings.Split(v1, ".")
	p2 := strings.Split(v2, ".")

	maxLen := len(p1)
	if len(p2) > maxLen {
		maxLen = len(p2)
	}

	for i := 0; i < maxLen; i++ {
		n1, n2 := 0, 0

		if i < len(p1) {
			fmt.Sscanf(p1[i], "%d", &n1)
		}
		if i < len(p2) {
			fmt.Sscanf(p2[i], "%d", &n2)
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	return 0
}

// Cleanup 清理临时目录
func (g *GitClient) Cleanup(tempPath string) {
	if tempPath != "" {
		os.RemoveAll(tempPath)
	}
}
