# SkillHub Project Setup Guide

## 1. 初始化 Wails + Vue 3 项目

```bash
# 安装 Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 创建项目（使用 Vue + TypeScript 模板）
wails init -n skillhub -t vue-ts

# 进入项目
cd skillhub
```

## 2. 安装前端依赖

```bash
cd frontend

# 安装 Tailwind CSS
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p

# 安装其他依赖
npm install axios vue-router@4 pinia

cd ..
```

## 3. 项目结构

```
skillhub/
├── backend/                    # Go 后端代码
│   ├── main.go                # 入口文件
│   ├── app.go                 # Wails 应用主逻辑
│   ├── config.go              # 配置管理
│   ├── skill/
│   │   ├── manager.go         # Skill 管理器
│   │   ├── storage.go         # 文件系统操作
│   │   ├── types.go           # Skill 类型定义
│   │   └── sync.go            # 数据同步（GitHub 备份）
│   ├── tools/
│   │   ├── interface.go       # 工具适配器接口
│   │   ├── base.go            # 适配器基类（Enable/Disable）
│   │   ├── adapters.go        # 各工具适配器实现
│   │   └── registry.go        # 适配器注册表
│   └── utils/
│       ├── git.go             # Git 操作
│       ├── file.go            # 文件操作（复制、解压）
│       ├── security.go        # 路径安全校验
│       ├── symlink.go         # 目录链接（跨平台API）
│       ├── symlink_windows.go # Windows Junction
│       ├── symlink_unix.go    # Unix Symlink
│       ├── helpers.go         # 辅助函数
│       └── skills.go          # Skill 工具函数
│
├── frontend/                   # Vue 3 前端
│   ├── src/
│   │   ├── main.ts            # Vue 入口
│   │   ├── App.vue            # 根组件
│   │   ├── router/            # 路由配置
│   │   ├── stores/            # Pinia 状态管理
│   │   ├── components/        # 公共组件
│   │   └── views/             # 页面组件
│   └── wailsjs/               # Wails 生成的绑定
│
├── wails.json                 # Wails 配置
└── go.mod                     # Go 依赖
```

## 4. 配置文件

### wails.json
```json
{
  "$schema": "https://wails.io/schemas/config.v2.json",
  "name": "skillhub",
  "outputfilename": "skillhub",
  "frontend": {
    "dir": "./frontend",
    "install": "npm install",
    "build": "npm run build",
    "dev": "npm run dev"
  },
  "author": {
    "name": "Your Name",
    "email": "your.email@example.com"
  }
}
```

### frontend/tailwind.config.js
```javascript
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        cyber: {
          dark: '#0a0a0f',
          panel: '#12121a',
          border: '#1e1e2e',
          accent: '#00d4aa',
          accent2: '#a855f7',
        }
      },
      fontFamily: {
        mono: ['JetBrains Mono', 'monospace'],
        sans: ['Inter', 'system-ui', 'sans-serif'],
      }
    },
  },
  plugins: [],
}
```

## 5. 启动开发服务器

```bash
# 在项目根目录
wails dev

# 这会同时启动：
# - Go 后端（自动热重载）
# - Vue 前端开发服务器
# - 桌面应用窗口
```

## 6. 构建生产版本

```bash
# Windows
wails build -platform windows

# macOS
wails build -platform darwin

# Linux
wails build -platform linux

# 所有平台
wails build
```
