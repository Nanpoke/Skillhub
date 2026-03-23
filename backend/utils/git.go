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
	"time"
)

// GitHubRelease GitHub Release API 响应
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// GitHubCommit GitHub Commit API 响应
type GitHubCommit struct {
	Commit struct {
		Committer struct {
			Date time.Time `json:"date"`
		} `json:"committer"`
	} `json:"commit"`
}

// GitClient Git 操作客户端
type GitClient struct {
	GitHubToken string // GitHub API Token，可选，用于解除限流
}

// NewGitClient 创建 Git 客户端
func NewGitClient(githubToken string) *GitClient {
	return &GitClient{
		GitHubToken: githubToken,
	}
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

// allowedGitDomains 允许的 Git 仓库域名
var allowedGitDomains = []string{
	"https://github.com/",
	"https://gitlab.com/",
	"https://gitee.com/",
}

// ValidateGitURL 验证 Git URL 是否来自允许的域名
func ValidateGitURL(url string) error {
	url = strings.TrimSpace(url)

	for _, prefix := range allowedGitDomains {
		if strings.HasPrefix(url, prefix) {
			return nil
		}
	}

	return fmt.Errorf("URL 必须来自 github.com、gitlab.com 或 gitee.com")
}

// Clone 浅克隆仓库到临时目录，支持子路径稀疏克隆
func (g *GitClient) Clone(url string, subPath string) (*CloneResult, error) {
	if !g.IsGitInstalled() {
		return nil, fmt.Errorf("Git 未安装，请先安装 Git")
	}

	// 验证 URL 来源
	if err := ValidateGitURL(url); err != nil {
		return nil, err
	}

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "skillhub-clone-*")
	if err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}

	var output []byte
	if subPath != "" {
		// 使用稀疏克隆只下载指定子路径
		cmd := exec.Command("git", "clone", "--filter=blob:none", "--sparse", url, tempDir)
		output, err = cmd.CombinedOutput()
		if err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("稀疏克隆失败: %s\n%s", err.Error(), string(output))
		}

		// 设置稀疏检出路径
		cmd = exec.Command("git", "sparse-checkout", "set", subPath)
		cmd.Dir = tempDir
		output, err = cmd.CombinedOutput()
		if err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("设置稀疏检出路径失败: %s\n%s", err.Error(), string(output))
		}

		// 拉取代码
		cmd = exec.Command("git", "pull")
		cmd.Dir = tempDir
		output, err = cmd.CombinedOutput()
		if err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("拉取代码失败: %s\n%s", err.Error(), string(output))
		}
	} else {
		// 普通浅克隆整个仓库
		cmd := exec.Command("git", "clone", "--depth", "1", url, tempDir)
		output, err = cmd.CombinedOutput()
		if err != nil {
			os.RemoveAll(tempDir)
			return nil, fmt.Errorf("克隆失败: %s\n%s", err.Error(), string(output))
		}
	}

	// 扫描 Skill 目录（传入子路径进行过滤）
	skills := g.ScanSkills(tempDir, subPath)

	return &CloneResult{
		TempPath: tempDir,
		Skills:   skills,
	}, nil
}

// ScanSkills 扫描目录中的 Skills，支持子路径过滤
func (g *GitClient) ScanSkills(dir string, subPath string) []string {
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

			// 如果指定了子路径，只返回该子路径下的Skill
			if subPath != "" {
				// 标准化路径分隔符
				normalizedRelPath := filepath.ToSlash(relPath)
				normalizedSubPath := filepath.ToSlash(subPath)

				// 检查是否在子路径下（精确匹配或以子路径+斜杠开头，避免前缀误匹配）
				pathMatches := normalizedRelPath == normalizedSubPath ||
					strings.HasPrefix(normalizedRelPath, normalizedSubPath + "/")

				if !pathMatches {
					return nil
				}

				// 始终返回相对于仓库根目录的完整路径
				if relPath == "." {
					skills = append(skills, "")
				} else {
					skills = append(skills, relPath)
				}
			} else {
				// 没有子路径过滤，返回所有相对路径
				if relPath == "." {
					// 根目录就是 Skill
					skills = append(skills, "")
				} else {
					skills = append(skills, relPath)
				}
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

// PullWithGitDir 拉取更新（支持分离的 .git 目录）
// gitDir: .git 目录的路径
// workTree: 工作目录的路径
func (g *GitClient) PullWithGitDir(gitDir, workTree string) error {
	cmd := exec.Command("git", "--git-dir", gitDir, "--work-tree", workTree, "pull")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("拉取失败: %s", string(output))
	}
	return nil
}

// GetTagWithGitDir 获取 Git tag（支持分离的 .git 目录）
func (g *GitClient) GetTagWithGitDir(gitDir string) (string, error) {
	cmd := exec.Command("git", "--git-dir", gitDir, "describe", "--tags", "--abbrev=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 如果没有tag，返回unknown而不是报错
		if strings.Contains(string(output), "No names found") {
			return "unknown", nil
		}
		return "", fmt.Errorf("获取 tag 失败: %s", string(output))
	}
	return strings.TrimSpace(string(output)), nil
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

// MoveGitDir 复制 .git 目录（支持跨磁盘）
// srcPath: 包含 .git 子目录的源路径（仓库根目录）
// destPath: .git 目录的目标路径（直接是 .git 内容存放的位置）
// 注意：改为复制模式，不删除源目录，以支持同一仓库安装多个 Skill
func (g *GitClient) MoveGitDir(srcPath, destPath string) error {
	srcGit := filepath.Join(srcPath, ".git")
	destGit := destPath // 直接使用 destPath，不再追加 .git

	// 检查源 .git 是否存在
	srcInfo, err := os.Stat(srcGit)
	if err != nil {
		if os.IsNotExist(err) {
			// .git 不存在，可能是用户没有安装 Git 或克隆失败
			// 返回 nil 允许继续（但不会保存 .git 用于更新检查）
			return nil
		}
		return fmt.Errorf("无法访问 .git 目录: %w", err)
	}

	// 确保目标父目录存在
	if err := os.MkdirAll(filepath.Dir(destGit), 0755); err != nil {
		return err
	}

	// 检查 .git 是文件还是目录（Git worktree 可能使用文件）
	if !srcInfo.IsDir() {
		// 如果是文件（gitfile），读取内容获取实际 .git 路径
		content, err := os.ReadFile(srcGit)
		if err != nil {
			return fmt.Errorf("读取 .git 文件失败: %w", err)
		}
		// gitfile 格式: gitdir: /path/to/.git/worktrees/xxx
		// 直接复制这个文件到目标位置
		return os.WriteFile(destGit, content, srcInfo.Mode())
	}

	// 复制 .git 目录（不删除源目录，支持同一仓库安装多个 Skill）
	if err := CopyDir(srcGit, destGit, false); err != nil {
		return fmt.Errorf("复制 .git 目录失败: %w", err)
	}

	return nil
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

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 如果配置了GitHub Token，添加到请求头
	if g.GitHubToken != "" {
		req.Header.Set("Authorization", "token "+g.GitHubToken)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
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

// FetchLatestCommitTime 从 GitHub API 获取指定路径的最新提交时间
func (g *GitClient) FetchLatestCommitTime(owner, repo, path string) (time.Time, error) {
	// 首先尝试最常见的两种路径，减少API调用
	tryPaths := []string{
		path,          // 直接路径
		"skills/" + path, // 最常见的skills目录前缀
	}

	for _, tryPath := range tryPaths {
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits?path=%s&per_page=1", owner, repo, tryPath)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return time.Time{}, fmt.Errorf("创建请求失败: %w", err)
		}

		// 如果配置了GitHub Token，添加到请求头
		if g.GitHubToken != "" {
			req.Header.Set("Authorization", "token "+g.GitHubToken)
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return time.Time{}, fmt.Errorf("请求 GitHub API 失败: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 403 || resp.StatusCode == 429 {
			return time.Time{}, fmt.Errorf("GitHub API 限流，请稍后再试或配置 GitHub Token")
		}

		if resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			var commits []GitHubCommit
			if err := json.Unmarshal(body, &commits); err != nil {
				continue
			}

			if len(commits) > 0 {
				return commits[0].Commit.Committer.Date, nil
			}
		}
	}

	// 所有尝试都失败
	return time.Time{}, fmt.Errorf("未找到提交记录")
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
