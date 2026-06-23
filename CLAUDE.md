# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SkillHub is a Wails v2 desktop application for managing AI coding tool Skills across multiple platforms (Claude Code, OpenCode, Cursor, CodeBuddy, Trae, and custom tools). One storage location, synced to all tools via copy-on-enable / delete-on-disable.

## Build & Dev Commands

```bash
# Dev mode (hot reload for both frontend and backend)
wails dev

# Build production binary
wails build

# Frontend only (without Wails)
cd frontend && npm run dev

# Frontend type check + build
cd frontend && npm run build
```

No test framework is configured. No linting is set up.

## Architecture

**Stack**: Go 1.23 (backend) + Wails v2 + Vue 3 + TypeScript + Tailwind CSS v4 + Pinia

**Communication**: Frontend calls Go backend via Wails auto-generated bindings in `frontend/wailsjs/go/backend/App`. Every public method on `backend.App` becomes callable as `App.MethodName()` in TypeScript. After adding/modifying Go methods, run `wails dev` or `wails generate module` to regenerate bindings.

### Backend (`backend/`)

```
app.go              → App struct (Wails-bound). All methods exposed to frontend live here.
                       Delegates to skill.Manager for business logic.
skill/
  types.go          → Data types: Skill, Metadata, InstallOptions, SkillInfo, AppSettings, SyncConfig, etc.
  manager.go        → Manager struct. Core logic: install, delete, toggle, scan, update check,
                       Git clone, category management, symlink-based enable/disable. Entry point.
  storage.go        → Storage struct. File system I/O: metadata JSON persistence, directory
                       management, operation logs, settings, custom tools.
  sync.go           → Data sync (Git-based backup). InitSync, SyncPush, SyncPull, GetSyncStatus,
                       RemoveSync, HasPendingChanges. Embeds token in URL, nested-repo-safe add.
tools/
  interface.go      → Adapter interface (ID, Name, SkillsPath, IsInstalled, EnableSkill, DisableSkill, etc.)
  base.go           → BaseAdapter. Shared implementation. Enable = CopyDir, Disable = RemoveAll.
  adapters.go       → Concrete adapters: Claude, OpenCode, Cursor, CodeBuddy, Trae.
                       Each knows its tool's skills path on disk.
  registry.go       → Registry. Thread-safe map of Adapter instances. Also registers custom tools.
utils/
  git.go            → GitClient. Clone, pull, tag operations, GitHub API (releases, commits),
                       URL parsing, version comparison. All git commands use HideWindow on Windows.
  security.go       → Path traversal prevention (SanitizeZipPath, ValidateSkillName, ValidatePathInDir).
  file.go           → CopyDir, CopyFile, UnzipFile, GetHomeDir, ExpandPath.
  helpers.go, skills.go → Misc utilities.
  symlink.go        → CreateSkillLink, IsSkillLink, RemoveSkillLink (platform-agnostic API).
  symlink_windows.go → Windows: creates Junction via `mklink /J`.
  symlink_unix.go   → Unix: creates os.Symlink.
```

### Frontend (`frontend/src/`)

```
App.vue             → Root component. Global notification system, confirm dialog, EventsOn for update checks.
router/index.ts     → Hash-history routes: /, /settings, /install, /wizard, /viewer/:id, /history, /import-export
stores/
  skills.ts         → Pinia store. Skills list, filtering (category/tags/tools/search/update-only),
                       activeFilterCount, clearAllFilters, batch update, ignore toggle. Calls App.* bindings.
  settings.ts       → Pinia store. App settings, theme, first-run state.
views/              → Page components matching routes.
components/         → Shared: ConfirmDialog.vue.
```

### Data Storage

Default data path: `~/.skill-hub/` (configurable on first run via wizard).

A "pointer" config is also saved at `%APPDATA%\SkillHub` (Windows) or `~/.config/skillhub` (Unix) so the app can find the data path on subsequent launches.

```
~/.skill-hub/
├── skills/              → Skill file copies (author-skillname dirs)
├── config/              → Per-tool enable config (claude-code.json, etc.)
├── metadata/            → Per-skill JSON metadata (author-skillname.json)
├── git/                 → Detached .git dirs for git-installed skills (used for update checks)
├── history/             → Operation logs (operations.log, auto-cleaned after 10 days)
├── settings.json        → App settings (path, theme, GitHub token, custom categories, Sync config)
├── custom-tools.json    → User-added custom tool definitions
├── .gitignore           → Auto-created by InitSync, excludes config/history/settings
└── .git/                → Created by InitSync when data sync is configured
```

### Key Design Decisions

- **Skill ID format**: `{author}-{skillname}` where author comes from Git URL owner (for git installs) or SKILL.md frontmatter/LICENSE.txt parsing (for local installs).
- **Enable/Disable mechanism**: Enable = `CopyDir` from `~/.skill-hub/skills/{id}` to tool's skills path (or `CreateSkillLink` Junction/symlink for install from Sync pull). Disable = `RemoveSkillLink` (detects if it's a symlink → remove link only; otherwise → RemoveAll for legacy copies).
- **Detached .git**: Git-installed skills store their .git separately under `~/.skill-hub/git/{id}`. This allows updating skills via `git --git-dir ... --work-tree ... pull` without mixing .git into the skill content directory.
- **SourceURL exposure**: `Manager.GetSkill` returns `SourceURL: ""` for `SourceTypeLocal` skills, even though `Metadata.SourceURL` may hold the local file path on disk. This keeps the frontend's "Skill 来源" UI from rendering a local path as a clickable URL; the UI instead shows "本地导入" for locally-installed skills.
- **Dual update detection**: Tag version comparison first; if no Release/tag exists, falls back to GitHub API commit timestamp comparison.
- **Security**: Git URL domain whitelist (github.com, gitlab.com, gitee.com), path traversal validation on skill names and zip extraction, null byte checks.

## Conventions

- Go module name: `skillhub` (not `github.com/...`)
- Frontend uses hash-based routing (`createWebHashHistory`) — required by Wails webview
- All Git CLI commands on Windows use `syscall.SysProcAttr{HideWindow: true}` to suppress console windows
- Chinese UI strings are inline (no i18n framework)
- Wails auto-generated files in `frontend/wailsjs/` should not be manually edited
