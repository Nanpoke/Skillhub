# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased] - 2026-03-23

### Added

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
