package skill

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// SyncStatus 同步状态
type SyncStatus struct {
	IsConfigured bool   `json:"is_configured"`
	RemoteURL    string `json:"remote_url"`
	Branch       string `json:"branch"`
	LastPushAt   string `json:"last_push_at"`
	LastPullAt   string `json:"last_pull_at"`
	HasChanges   bool   `json:"has_changes"`
}

// InitSync 初始化同步仓库
func (m *Manager) InitSync(remoteURL, branch, githubToken string) error {
	basePath := m.storage.GetBasePath()

	// 清理残留锁文件
	if err := cleanGitLocks(basePath); err != nil {
		return err
	}

	// 1. 检查 Git 是否安装
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("Git 未安装，请先安装 Git")
	}

	if branch == "" {
		branch = "main"
	}

	// 2. git init（如果尚未初始化）
	gitDir := filepath.Join(basePath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := gitCmd(ctx, basePath, "init")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git init 失败: %s\n%s", err.Error(), string(output))
		}
		// 强制将默认分支重命名为目标分支（git init 可能创建 master 而非 main）
		ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel2()
		cmd = gitCmd(ctx2, basePath, "branch", "-M", branch)
		cmd.CombinedOutput()
	}

	// 3. 写 .gitignore
	gitignorePath := filepath.Join(basePath, ".gitignore")
	gitignoreContent := `# SkillHub 同步排除
config/
history/
settings.json
custom-tools.json
`
	if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
		return fmt.Errorf("写入 .gitignore 失败: %w", err)
	}

	// 4. git remote add origin（已存在则 set-url）
	{
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := gitCmd(ctx, basePath, "remote", "add", "origin", remoteURL)
		if output, err := cmd.CombinedOutput(); err != nil {
			// 如果 remote 已存在，更新 URL
			ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel2()
			cmd = gitCmd(ctx2, basePath, "remote", "set-url", "origin", remoteURL)
			output, err = cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("设置远程仓库失败: %s\n%s", err.Error(), string(output))
			}
		}
	}

	// 5. 保存 remoteURL、branch 到 settings.json
	settings, err := m.storage.LoadSettings()
	if err != nil {
		return fmt.Errorf("加载设置失败: %w", err)
	}

	if branch == "" {
		branch = "main"
	}
	settings.Sync = &SyncConfig{
		RemoteURL: remoteURL,
		Branch:    branch,
	}

	if err := m.storage.SaveSettings(settings); err != nil {
		return fmt.Errorf("保存设置失败: %w", err)
	}

	return nil
}

// SyncPush 推送本地变更到远程
func (m *Manager) SyncPush(authorName, message, githubToken string) error {
	basePath := m.storage.GetBasePath()

	// 清理残留锁文件（上次崩溃可能遗留）
	if err := cleanGitLocks(basePath); err != nil {
		return err
	}

	// 检查是否已初始化
	gitDir := filepath.Join(basePath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("同步未初始化，请先配置远程仓库")
	}

	settings, err := m.storage.LoadSettings()
	if err != nil || settings.Sync == nil {
		return fmt.Errorf("同步配置丢失，请重新初始化")
	}

	// 推送前检查 Token
	if githubToken == "" {
		return fmt.Errorf("未配置 GitHub Token，请在设置页面中配置 GitHub Token（创建时需勾选 repo 权限）")
	}

	branch := settings.Sync.Branch
	if branch == "" {
		branch = "main"
	}
	remoteURL := settings.Sync.RemoteURL

	// 批量收集需添加的文件，绕过 Git 2.25+ 对嵌套仓库的 gitlink 处理
	var filesToAdd []string
	for _, dir := range []string{"skills", "git", "metadata"} {
		filepath.Walk(filepath.Join(basePath, dir), func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			relPath, err := filepath.Rel(basePath, path)
			if err != nil {
				return nil
			}
			filesToAdd = append(filesToAdd, filepath.ToSlash(relPath))
			return nil
		})
	}

	// 分批 git add（每批最多 100 个文件）
	const batchSize = 100
	for i := 0; i < len(filesToAdd); i += batchSize {
		end := i + batchSize
		if end > len(filesToAdd) {
			end = len(filesToAdd)
		}
		batch := filesToAdd[i:end]
		if len(batch) == 0 {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		cmd := gitCmd(ctx, basePath, append([]string{"add"}, batch...)...)
		cmd.CombinedOutput()
		cancel()
	}

	// 暂存已追踪文件的删除和修改（git add -u 不影响未追踪文件，不会触发 gitlink 问题）
	{
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := gitCmd(ctx, basePath, "add", "-u", "skills/", "git/", "metadata/")
		cmd.CombinedOutput()
	}

	// 也添加 .gitignore（如果它是新创建的）
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := gitCmd(ctx, basePath, "add", ".gitignore")
		cmd.CombinedOutput()
	}

	// 确保本地分支名与配置一致（git init 可能创建 master 而非 main）
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := gitCmd(ctx, basePath, "branch", "-M", branch)
		cmd.CombinedOutput()
	}

	// 2. git commit
	if message == "" {
		message = "Sync from SkillHub"
	}
	if authorName != "" {
		message = fmt.Sprintf("Sync from %s", authorName)
	}

	{
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := gitCmd(ctx, basePath, "commit", "-m", message, "--allow-empty")
		output, err := cmd.CombinedOutput()
		if err != nil {
			// "nothing to commit" 不算错误
			if !strings.Contains(string(output), "nothing to commit") {
				return fmt.Errorf("提交失败: %s\n%s", err.Error(), string(output))
			}
		}
	}

	// 3. 先拉取远程更新再推送（避免 non-fast-forward 被拒）
	// -c credential.helper= 禁用凭证助手，防止 Windows Credential Manager 中的旧凭证覆盖 URL 中的 Token
	pullURL := embedTokenInURL(remoteURL, githubToken)
	{
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		cmd := gitCmd(ctx, basePath, "-c", "credential.helper=", "fetch", pullURL, branch)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// fetch 失败（新仓库可能还没有分支），不阻止推送
			// 继续尝试推送
		} else {
			// 合并远程变更，冲突时以本地为准
			ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel2()
			cmd := gitCmd(ctx2, basePath, "merge", "FETCH_HEAD", "-X", "ours", "--no-edit")
			cmd.CombinedOutput() // 合并失败不阻止推送
			_ = output // suppress unused warning
		}
	}

	// 4. git push（使用 token 认证，禁用凭证助手）
	pushURL := embedTokenInURL(remoteURL, githubToken)
	{
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		cmd := gitCmd(ctx, basePath, "-c", "credential.helper=", "push", pushURL, branch)
		output, err := cmd.CombinedOutput()
		if err != nil {
			errStr := string(output)
			if strings.Contains(errStr, "403") || strings.Contains(errStr, "Write access") {
				return fmt.Errorf(
					"推送失败：没有仓库写入权限。\n\n"+
						"请检查：\n"+
						"1. GitHub Token 是否包含 repo 权限（创建 Token 时需要勾选）\n"+
						"2. 仓库地址是否正确\n"+
						"3. 你的账号是否有该仓库的写入权限\n\n"+
						"获取 Token: https://github.com/settings/tokens")
			}
			if strings.Contains(errStr, "context deadline exceeded") {
				return fmt.Errorf("推送超时：网络连接较慢或被中断，请检查网络后重试")
			}
			return fmt.Errorf("推送失败: %s\n%s", err.Error(), errStr)
		}
	}

	// 5. 更新最后推送时间
	settings.Sync.LastPushAt = time.Now().Format(time.RFC3339)
	m.storage.SaveSettings(settings)

	return nil
}

// SyncPull 从远程拉取更新
func (m *Manager) SyncPull(authorName, githubToken string) error {
	basePath := m.storage.GetBasePath()

	// 清理残留锁文件
	if err := cleanGitLocks(basePath); err != nil {
		return err
	}

	// 检查是否已初始化
	gitDir := filepath.Join(basePath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("同步未初始化，请先配置远程仓库")
	}

	settings, err := m.storage.LoadSettings()
	if err != nil || settings.Sync == nil {
		return fmt.Errorf("同步配置丢失，请重新初始化")
	}

	branch := settings.Sync.Branch
	if branch == "" {
		branch = "main"
	}
	remoteURL := settings.Sync.RemoteURL

	// 1. git fetch（使用 token 认证，禁用凭证助手）
	pullURL := embedTokenInURL(remoteURL, githubToken)
	{
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		cmd := gitCmd(ctx, basePath, "-c", "credential.helper=", "fetch", pullURL, branch)
		output, err := cmd.CombinedOutput()
		if err != nil {
			if strings.Contains(string(output), "context deadline exceeded") {
				return fmt.Errorf("拉取超时：网络连接较慢或被中断，请检查网络后重试")
			}
			return fmt.Errorf("拉取失败: %s\n%s", err.Error(), string(output))
		}
	}

	// 2. git merge（策略：ours 优先，本地优先）
	{
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := gitCmd(ctx, basePath, "merge", "FETCH_HEAD", "-X", "ours", "--no-edit")
		output, err := cmd.CombinedOutput()
		if err != nil {
			// 冲突时记录但不阻止
			m.logOperation("sync-pull", "merge conflict", "warning")
			return fmt.Errorf("合并失败，可能存在冲突: %s\n%s", err.Error(), string(output))
		}
	}

	// 3. 更新最后拉取时间
	settings.Sync.LastPullAt = time.Now().Format(time.RFC3339)
	m.storage.SaveSettings(settings)

	m.logOperation("sync-pull", authorName, "success")
	return nil
}

// GetSyncStatus 获取同步状态
func (m *Manager) GetSyncStatus() (*SyncStatus, error) {
	basePath := m.storage.GetBasePath()

	gitDir := filepath.Join(basePath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return &SyncStatus{IsConfigured: false}, nil
	}

	settings, err := m.storage.LoadSettings()
	if err != nil || settings.Sync == nil {
		return &SyncStatus{IsConfigured: false}, nil
	}

	hasChanges, _ := m.HasPendingChanges()

	return &SyncStatus{
		IsConfigured: true,
		RemoteURL:    settings.Sync.RemoteURL,
		Branch:       settings.Sync.Branch,
		LastPushAt:   settings.Sync.LastPushAt,
		LastPullAt:   settings.Sync.LastPullAt,
		HasChanges:   hasChanges,
	}, nil
}

// RemoveSync 移除同步配置，保留本地数据
func (m *Manager) RemoveSync() error {
	basePath := m.storage.GetBasePath()

	// 移除 remote
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := gitCmd(ctx, basePath, "remote", "remove", "origin")
	cmd.CombinedOutput() // 忽略错误

	// 清除设置中的 sync 配置
	settings, err := m.storage.LoadSettings()
	if err == nil {
		settings.Sync = nil
		m.storage.SaveSettings(settings)
	}

	return nil
}

// HasPendingChanges 检查是否有待提交的变更
func (m *Manager) HasPendingChanges() (bool, error) {
	basePath := m.storage.GetBasePath()

	gitDir := filepath.Join(basePath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := gitCmd(ctx, basePath, "status", "--porcelain", "skills/", "git/", "metadata/")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(output)) != "", nil
}

// cleanGitLocks 清理残留的 git 锁文件
// 返回 error 表示锁文件存在且无法清理
func cleanGitLocks(basePath string) error {
	lockFiles := []string{
		filepath.Join(basePath, ".git", "index.lock"),
		filepath.Join(basePath, ".git", "HEAD.lock"),
	}
	for _, f := range lockFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			continue
		}
		// 先尝试直接删除
		if err := os.Remove(f); err != nil {
			// 删除失败（可能被同步软件锁定），尝试重命名绕开
			tmpName := f + ".stale-" + time.Now().Format("20060102-150405")
			if renameErr := os.Rename(f, tmpName); renameErr != nil {
				return fmt.Errorf("无法清理残留锁文件 %s（可能被同步软件占用），请手动删除该文件后重试", f)
			}
		}
	}
	return nil
}

// embedTokenInURL 将 GitHub token 嵌入 URL（用于认证）
func embedTokenInURL(url, token string) string {
	if token == "" || !strings.HasPrefix(url, "https://") {
		return url
	}
	// https://github.com/owner/repo -> https://token@github.com/owner/repo
	return strings.Replace(url, "https://", "https://"+token+"@", 1)
}

// gitCmd 创建带超时和凭证保护的基础 exec.Cmd
// 禁用 Git 交互提示，防止在 HideWindow 模式下挂起
func gitCmd(ctx context.Context, basePath string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = basePath
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GCM_INTERACTIVE=never",
	)
	return cmd
}
