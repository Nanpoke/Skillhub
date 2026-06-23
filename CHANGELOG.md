# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added

#### Skill 来源 URL 可点击打开

- 后端新增 `OpenURL(url string) error` 绑定方法（`backend/app.go`），调用 `runtime.BrowserOpenURL` 在系统默认浏览器中打开 URL；仅放行 `http://` / `https://` 协议，拒绝 `javascript:` / `file:` 等
- 主窗口 Skill 卡片展开后的"Skill 描述"区块改名为"Skill 来源"；Git 来源 Skill 显示完整 URL + 外链图标（`fa-external-link-alt`），点击调用 `App.OpenURL` 在系统浏览器中打开
- 本地导入 Skill（`SourceTypeLocal`）在"Skill 来源"区块显示"📂 本地导入"占位文字（不可点击）

#### 搜索匹配备注

- 主窗口顶部搜索框的匹配字段从 4 个（`name` / `description` / `tags` / `author`）扩展为 5 个，新增 `notes`（用户填写的备注）

### Fixed

- `GetSkill` 序列化时对 `SourceTypeLocal` 类型的 Skill 清空 `SourceURL` 字段：之前本地安装路径（如 `C:\...\xxx.zip`）会作为 `source_url` 暴露给前端，触发"打开来源失败"通知；现在本地导入 Skill 在 UI 上自然走"本地导入"分支

## [1.4] - 2026-06-21

### Added

#### 数据同步（GitHub 备份）

- 新增 GitHub 数据同步功能，支持将 Skills 数据备份到远程仓库
- **InitSync**：初始化同步仓库，创建 .gitignore（排除 config/history/settings），配置 remote
- **SyncPush**：推送本地 skills/metadata/git 变更到远程，分批 git add 绕过嵌套仓库检测，自动 pull 再 push
- **SyncPull**：从远程拉取并合并变更，冲突策略 ours（本地优先）
- **GetSyncStatus**：查询同步状态（是否已配置、远程地址、分支、最后推送/拉取时间、是否有待提交变更）
- **RemoveSync**：断开同步，移除 remote 配置，保留本地数据
- **锁文件清理**：推送/拉取前自动清理残留的 index.lock / HEAD.lock
- 设置页面完整同步管理 UI（初始化、推送弹窗、拉取确认、断开、变更远程地址）
- 推送需 GitHub Token 且必须带有 repo 权限
- Token 嵌入 URL 进行认证，禁用 Git 凭证助手防止旧凭证覆盖

#### Symlink/Junction 启禁用机制

- `utils/symlink.go`：跨平台 API（CreateSkillLink / IsSkillLink / RemoveSkillLink）
- Windows 实现：`mklink /J` junction
- Unix 实现：`os.Symlink`
- 删除时自动识别链接类型：是链接则只删链接节点，非链接（旧版复制目录）则 RemoveAll
- 自定义工具删除改用 RemoveSkillLink，防止误删源目录

### Changed

- Trae 工具的默认 Skills 路径从 `~/.trae/skills` 改为 `~/.trae-cn/skills`

---

## [1.3] - 2026-05-05

### Fixed

- **Sync Push 嵌套仓库推送修复**：`git/` 目录（裸 Git 仓库）和 `skills/` 目录曾被 Git 2.25+ 当作子模块引用（gitlink）处理，导致推送到远程后文件内容丢失。改为对所有同步目录（skills/、git/、metadata/）逐文件 `git add`，彻底绕过嵌套仓库检测。

### Added

#### 筛选状态条（Filter Bar）

- 主界面新增 Filter Bar，仅在有活跃筛选条件时显示
- 左侧以 chip 标签展示当前所有活跃筛选条件（分类、标签、工具、搜索词、可更新标记）
- 每个 chip 可点击 × 单独移除对应筛选条件
- 右侧"清除全部"按钮，一键归零所有筛选
- 无筛选条件时 Filter Bar 隐藏不占空间

#### Skill更新机制优化

**问题解决**：
- 解决无Release和tag的Skill无法检测更新的问题
- 大部分GitHub Skill没有发布Release也没有tag，导致无法检查更新

**实现方案**：
- **双重更新检测机制**：优先使用tag版本比较，无tag或无Release时回退到提交时间比较
- **子路径支持**：支持Skill在Git仓库子目录中的更新检查
- **GitHub API集成**：新增`FetchLatestCommitTime`方法获取指定路径的最新提交时间
- **智能路径匹配**：自动尝试直接路径和`skills/`前缀两种常见格式
- **向后兼容**：旧数据自动解析并保存SubPath，无需数据迁移
- **错误处理增强**：API调用失败时静默跳过，不影响用户体验

**技术实现**：
- `Metadata`结构新增`SubPath`字段，保存Skill在Git仓库中的子路径
- `GitClient`新增`FetchLatestCommitTime(owner, repo, path)`方法
- `InstallFromGit`方法自动解析并保存SubPath信息
- `CheckSkillUpdates`方法优化更新检查逻辑，支持tag和提交时间双模式

### Added

#### 更新功能配套优化

1. **空状态适配优化**
   - 筛选"只看可更新"且无结果时，显示定制空状态
   - 提示内容：🎉 太棒了！所有 Skill 都是最新版本
   - 提供"查看全部 Skill"按钮，点击后取消更新筛选

2. **批量更新功能**
   - 在统计栏可更新数旁边添加"一键更新"按钮
   - 仅当有可更新Skill时显示
   - 点击后了显示确认弹窗
   - 更新过程中显示进度提示（更新中 X/N）
   - 支持批量更新所有可更新的Skill

3. **更新忽略功能**
   - Skill卡片操作栏新增"忽略更新"按钮（铃铛图标）
   - 仅当Skill有更新时显示
   - 点击后切换忽略状态（已忽略/可提醒）
   - 被忽略的Skill不显示"有更新"标记
   - 被忽略的Skill不计入可更新数统计
   - 图标样式：已忽略时显示灰色铃铛，可提醒时显示橙色铃铛

4. **时间显示优化**
   - 时间显示逻辑：优先使用更新时间，其次使用安装时间
   - 鼠标hover时显示tooltip提示："更新时间"或"安装时间"

5. **更新状态优化**
   - 更新成功后直接修改本地状态，无需全量刷新
   - 自动清除"有更新"标记
   - 自动更新时间为当前时间
   - 避免列表跳动，提升体验

### Added

#### GitHub Token 配置功能

**后端**
- `AppSettings` 结构新增 `GitHubToken` 字段
- `GitClient` 结构新增 `GitHubToken` 属性
- 所有GitHub API请求自动带上 `Authorization: token` 请求头
- 所有Git客户端创建时自动读取配置的Token
- 限流错误提示优化，引导用户配置Token

**前端**
- 设置页设置"检查Skill更新"模块新增Token配置区域
- 密码输入框（输入Token）
- 保存按钮
- 说明文字和Token获取链接
- 页面加载时自动读取已保存的Token
- 保存成功/失败提示

### Changed

#### "有更新"标记设计优化

- **位置调整**：从标题内移到卡片容器右上角
- **绝对定位**：`-top-2 -right-2`，`z-20` 确保最上层
- **样式**：全圆角pill形状
- **背景渐变**：`linear-gradient(135deg, #00d4aa, #a855f7)`
- **脉冲动画**：2秒循环，呼吸效果
- 卡片容器添加 `position: relative`
- 卡片有更新时边框高亮和发光阴影

### Added

#### 更新按钮设计

- 新增主样式更新按钮，仅当有更新且未被忽略时显示
- 渐变背景，深色文字
- 固定正方形尺寸：36×36px
- 仅显示图标：`fa-arrow-up`
- Hover时向上浮动2px + 增强发光阴影
- Active时轻微缩放（scale 0.98）
- Disabled时半透明且禁用交互

### Changed

#### Trae工具路径

- Trae工具的默认Skills路径从 `~/.trae/skills` 改为 `~/.trae-cn/skills`
- 新安装的Skill在启用Trae时将同步到新路径
- 其他工具路径保持不变

### Added

#### 体验优化与功能完善（2026-03-23）

**用户体验优化**
1. **Git操作无窗口静默运行**
   - 所有Git命令调用添加 `HideWindow: true` 属性
   - Git安装和更新检查不再弹出黑色命令行窗口
   - 不影响任何功能逻辑，仅优化Windows平台体验

2. **Git安装动态加载提示**
   - 扫描Skill时显示动态步骤状态：连接仓库→下载内容→识别Skill→检查安装状态
   - 配合旋转加载动画，用户清晰感知执行进度
   - 解决用户误以为扫描无响应的问题

3. **主界面布局优化**
   - Skill详情页"备注"字段移到"同步到工具"下方
   - 备注文本框高度增加，适配常用标签气泡区域高度
   - 备注文字字号缩小为12px，界面更紧凑
   - 保存按钮位置调整到备注下方，操作逻辑更合理

4. **Git安装选择逻辑优化**
   - 未安装的Skills默认不选中，避免误安装
   - 自动识别已安装的Skills，显示绿色"已安装"标签且禁用选择
   - 显示"发现X个Skills，其中Y个已安装"提示信息
   - 从根源避免重复安装错误

5. **标签功能增强**
   - 安装页面标签输入框支持自动补全，显示已有标签列表
   - 安装页面添加常用标签气泡选择（按使用频率排序，前15个）
   - 主界面Skill详情页也添加常用标签气泡快速选择功能
   - 标签气泡样式适配暗黑主题，半透明cyber-panel背景

6. **界面布局调整**
   - 主界面侧边栏"工具筛选"模块移到"标签"模块上方
   - 侧边栏标签显示无数量限制，显示所有已存在的标签
   - 标签气泡点击添加逻辑优化，已添加的标签自动禁用

**Bug修复**
1. **Windows路径分隔符修复**
   - Git更新检查时路径自动转换为正斜杠格式
   - 解决Windows系统下GitHub API请求返回404的问题
   - 所有平台更新检查功能正常工作

---

## [1.0.0] - 2026-02-15

### Added
- Skill 浏览与管理主界面，支持卡片列表展示
- 本地安装 Skills（拖拽文件夹或 .zip 压缩包）
- Git 仓库安装 Skills（支持 user/repo 简写和完整 URL）
- 自定义分类和标签系统
- 为每个工具独立设置 Skill 启用/禁用
- SKILL.md 查看器（文件目录、渲染视图、代码视图）
- 操作历史记录（保留 10 天）
- 导入/导出配置功能
- 设置页面（外观、工具管理、存储、数据管理）
- 添加自定义 AI 工具支持
- 数据迁移功能
- 启动向导（首次运行配置）

### Supported AI Tools
- Claude Code
- OpenCode
- Cursor
- CodeBuddy
- Trae
- 自定义工具（用户可添加）

### Technical
- 后端: Go 1.20+ with Wails v2
- 前端: Vue 3 + TypeScript + Tailwind CSS
- 状态管理: Pinia
- 路由: Vue Router 4
