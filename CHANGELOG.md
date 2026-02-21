# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

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
