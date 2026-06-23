# SkillHub 开发任务清单（已归档）

> ⚠️ **此文件为项目初始开发阶段的任务规划，所有 Phase 1-8 均已实现。**
> 当前版本 v1.5，发布于 2026-06-23。
> 后续新增功能记录在 [CHANGELOG.md](CHANGELOG.md) 和 [README.md](README.md) 中。

## 📋 项目概览（历史参考）
- **技术栈**: Wails (Go) + Vue 3 + TypeScript + Tailwind CSS
- **目标**: 完整的 Skill 管理桌面应用
- **实际完成**: v1.0.0 于 2026-02-15。后续迭代 v1.0.1、v1.1、v1.2、v1.3、v1.4、v1.5

---

## Phase 1: 项目搭建 (已完成)

- [x] 安装 Go 1.20+
- [x] 安装 Wails CLI
- [x] 安装 Node.js 18+
- [x] 创建 Wails 项目
- [x] 配置 Tailwind CSS
- [x] 安装前端依赖
- [x] 配置路由和 Pinia
- [x] 创建代码目录结构

## Phase 2: 核心功能 - 配置管理 (已完成)

- [x] 配置结构体和管理
- [x] 首次启动向导
- [x] Wails 前端绑定

## Phase 3: 核心功能 - Skill 管理 (已完成)

- [x] Skill 类型定义和存储
- [x] 主界面（卡片列表、搜索、筛选、展开详情）
- [x] Pinia Store（加载、过滤、搜索）

## Phase 4: 核心功能 - 工具适配 (已完成)

- [x] ToolAdapter 接口（Detect/Enable/Disable）
- [x] 各工具适配器：Claude、OpenCode、Cursor、CodeBuddy、Trae
- [x] 前端启用/禁用同步界面

## Phase 5: 安装功能 (已完成)

- [x] Git 安装（3种 URL 格式、子路径支持）
- [x] 本地安装（拖拽/选择、.zip 解压）
- [x] skills.sh 浏览（可选功能，已实现）

## Phase 6: 辅助功能 (已完成)

- [x] SKILL.md 查看器（文件目录/渲染/代码三视图）
- [x] 元数据编辑（标签、分类、备注）
- [x] 历史记录（10天自动清理）

## Phase 7: 设置和数据 (已完成)

- [x] 设置页面（主题、更新检查、存储路径）
- [x] 路径迁移（完整迁移 7 项数据）
- [x] 导入/导出配置（.zip 格式）
- [x] 重置数据

## Phase 8: 优化和发布 (已完成)

- [x] Toast 通知、空状态、加载动画
- [x] 错误处理完善
- [x] 打包发布（v1.0.0）

---

## 📌 规划外新增功能（v1.1-v1.4）

以下功能是在初始开发计划之外逐步增加的，详见 CHANGELOG.md：

- **可更新数显示和筛选**（v1.1）
- **批量更新**（v1.1）
- **更新忽略功能**（v1.1）
- **双重更新检测机制**（tag 版本 + 提交时间，v1.2）
- **GitHub Token 配置**（v1.2）
- **Git 操作静默运行**（v1.3）
- **数据同步（GitHub 备份）**（v1.4）
- **Symlink/Junction 启禁用机制**（v1.4）
- **Filter Bar 筛选状态条**（v1.3）
- **Windows 兼容性修复**（v1.3）
- **Trae 路径从 `~/.trae/` 改为 `~/.trae-cn/`**（v1.4）
