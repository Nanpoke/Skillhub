<script setup lang="ts">
import { ref, inject } from 'vue'
import { useRouter } from 'vue-router'
import * as App from '../../wailsjs/go/backend/App'

const router = useRouter()

// 注入全局方法
const showNotification = inject<(message: string, type?: 'success' | 'error' | 'warning' | 'info', duration?: number) => void>('showNotification')

// 当前 Tab
const currentTab = ref<'export' | 'import'>('export')

// 导出状态
const exporting = ref(false)
const exportInfo = ref<{
  skillsCount: number
  gitCacheCount: number
  customToolsCount: number
  estimatedSize: string
} | null>(null)

// 导入状态
const importing = ref(false)
const importFile = ref<string>('')
const importPreview = ref<{
  version: string
  exportDate: string
  skillsCount: number
  gitCacheCount: number
  customToolsCount: number
} | null>(null)
const conflictMode = ref<'overwrite' | 'skip'>('overwrite')

// 加载导出信息
async function loadExportInfo() {
  try {
    const info = await App.GetExportInfo()
    exportInfo.value = {
      skillsCount: info.skills_count || 0,
      gitCacheCount: info.git_cache_count || 0,
      customToolsCount: info.custom_tools_count || 0,
      estimatedSize: info.estimated_size || '~0 MB'
    }
  } catch (e) {
    console.error('Failed to load export info:', e)
  }
}

// 切换 Tab
function switchTab(tab: 'export' | 'import') {
  currentTab.value = tab
  if (tab === 'export') {
    loadExportInfo()
  }
}

// 执行导出
async function exportData() {
  exporting.value = true
  try {
    const zipPath = await App.ExportData(true, true)
    showNotification?.(`导出成功！文件已保存至: ${zipPath}`, 'success', 5000)
  } catch (e) {
    showNotification?.('导出失败: ' + e, 'error')
  } finally {
    exporting.value = false
  }
}

// 选择导入文件
async function selectImportFile() {
  try {
    const filePath = await App.SelectFile()
    if (filePath) {
      importFile.value = filePath
      // 预览导入文件内容
      const preview = await App.PreviewImportFile(filePath)
      importPreview.value = {
        version: preview.version || '1.0',
        exportDate: preview.export_date || '',
        skillsCount: preview.skills_count || 0,
        gitCacheCount: preview.git_cache_count || 0,
        customToolsCount: preview.custom_tools_count || 0
      }
    }
  } catch (e) {
    showNotification?.('读取文件失败: ' + e, 'error')
    importFile.value = ''
    importPreview.value = null
  }
}

// 选择冲突处理模式
function selectConflictMode(mode: 'overwrite' | 'skip') {
  conflictMode.value = mode
}

// 执行导入
async function startImport() {
  if (!importFile.value) {
    showNotification?.('请先选择备份文件', 'warning')
    return
  }

  importing.value = true
  try {
    const mergeSkills = conflictMode.value === 'skip'
    await App.ImportData(importFile.value, true, mergeSkills)
    showNotification?.('导入成功！页面即将刷新...', 'success')
    setTimeout(() => {
      location.reload()
    }, 1500)
  } catch (e) {
    showNotification?.('导入失败: ' + e, 'error')
  } finally {
    importing.value = false
  }
}

// 返回
function goBack() {
  router.push('/settings')
}

// 初始化
loadExportInfo()
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-cyber-dark via-cyber-panel to-cyber-dark text-gray-300 font-sans">
    <!-- Header -->
    <header class="glass-panel glass-panel-no-hover sticky top-0 z-50 rounded-none border-b border-cyber-border">
      <div class="max-w-4xl mx-auto px-6 py-4 relative z-10">
        <div class="flex items-center gap-4">
          <button @click="goBack" class="btn-icon">
            <i class="fas fa-arrow-left text-sm"></i>
          </button>
          <div>
            <h1 class="text-xl font-bold font-mono">
              <span :class="currentTab === 'export' ? 'gradient-text' : 'text-white'">导出</span>
              <span class="text-gray-500 mx-1">/</span>
              <span :class="currentTab === 'import' ? 'gradient-text' : 'text-white'">导入</span>
            </h1>
            <p class="text-xs text-gray-500">备份和恢复 SkillHub 配置</p>
          </div>
        </div>
      </div>
    </header>

    <!-- Tab Navigation -->
    <div class="glass-panel glass-panel-no-hover rounded-none border-b border-cyber-border">
      <div class="max-w-4xl mx-auto px-6 relative z-10">
        <div class="flex gap-8">
          <button
            @click="switchTab('export')"
            :class="['tab-btn py-4 text-sm font-medium', currentTab === 'export' ? 'active' : 'text-gray-500']"
          >
            <i class="fas fa-file-export mr-2"></i>导出配置
          </button>
          <button
            @click="switchTab('import')"
            :class="['tab-btn py-4 text-sm font-medium', currentTab === 'import' ? 'active' : 'text-gray-500']"
          >
            <i class="fas fa-file-import mr-2"></i>导入配置
          </button>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <main class="flex-1 p-6">
      <div class="max-w-4xl mx-auto">

        <!-- Export Panel -->
        <div v-show="currentTab === 'export'" class="glass-panel rounded-2xl p-8 relative z-10">
          <div class="flex items-center gap-4 mb-6 relative z-10">
            <div class="w-14 h-14 rounded-xl bg-gradient-to-br from-cyber-accent/20 to-cyber-accent2/20 border border-cyber-accent/30 flex items-center justify-center">
              <i class="fas fa-file-export text-2xl text-cyber-accent"></i>
            </div>
            <div>
              <h2 class="text-lg font-semibold text-white">导出配置</h2>
              <p class="text-sm text-gray-500">将 SkillHub 完整配置导出为 .zip 备份文件</p>
            </div>
          </div>

          <!-- Export Content List -->
          <div class="mb-6 relative z-10">
            <label class="text-sm text-gray-400 block mb-3">导出内容</label>
            <div class="p-4 rounded-xl bg-cyber-dark border border-cyber-border">
              <ul class="space-y-3 text-sm">
                <li class="flex items-center justify-between">
                  <div class="flex items-center gap-2 text-gray-300">
                    <i class="fas fa-check text-cyber-accent text-xs w-4"></i>
                    Skills 文件本体及元数据
                  </div>
                  <span class="text-gray-500 font-mono text-xs">{{ exportInfo?.skillsCount || 0 }} 个</span>
                </li>
                <li class="flex items-center justify-between">
                  <div class="flex items-center gap-2 text-gray-300">
                    <i class="fas fa-check text-cyber-accent text-xs w-4"></i>
                    工具启用配置
                  </div>
                </li>
                <li class="flex items-center justify-between">
                  <div class="flex items-center gap-2 text-gray-300">
                    <i class="fas fa-check text-cyber-accent text-xs w-4"></i>
                    应用设置
                  </div>
                </li>
                <li class="flex items-center justify-between">
                  <div class="flex items-center gap-2 text-gray-300">
                    <i class="fas fa-check text-cyber-accent text-xs w-4"></i>
                    自定义工具配置
                  </div>
                  <span class="text-gray-500 font-mono text-xs">{{ exportInfo?.customToolsCount || 0 }} 个</span>
                </li>
                <li class="flex items-center justify-between">
                  <div class="flex items-center gap-2 text-gray-300">
                    <i class="fas fa-check text-cyber-accent text-xs w-4"></i>
                    Git 缓存
                  </div>
                  <span class="text-gray-500 font-mono text-xs">{{ exportInfo?.gitCacheCount || 0 }} 个仓库</span>
                </li>
                <li class="flex items-center gap-2 text-gray-500 pt-2 border-t border-cyber-border">
                  <i class="fas fa-times text-gray-600 text-xs w-4"></i>
                  <span class="line-through">历史记录</span>
                  <span class="text-xs text-gray-600">（不导出）</span>
                </li>
              </ul>
            </div>
          </div>

          <!-- File Structure Preview -->
          <div class="mb-6 relative z-10">
            <div class="flex items-center justify-between mb-3">
              <label class="text-sm text-gray-400">导出文件结构</label>
              <span class="text-xs text-gray-500">预计大小: {{ exportInfo?.estimatedSize || '~0 MB' }}</span>
            </div>
            <div class="p-4 rounded-xl bg-cyber-dark border border-cyber-border overflow-x-auto">
              <pre class="text-xs text-gray-400 font-mono leading-relaxed">skillhub-backup-{{ new Date().toISOString().split('T')[0] }}.zip
├── manifest.json          # 元数据清单
├── skills/                # Skill 文件本体
│   ├── vercel-react-best-practices/
│   └── frontend-design/
├── git/                   # Git 缓存
│   └── vercel-labs-agent-skills/
└── config/                # 工具启用配置</pre>
            </div>
          </div>

          <!-- Export Button -->
          <div class="relative z-10">
            <button
              @click="exportData"
              :disabled="exporting"
              class="btn-primary w-full text-lg flex items-center justify-center gap-2"
            >
              <i :class="['fas', exporting ? 'fa-spinner fa-spin' : 'fa-download']"></i>
              <span>{{ exporting ? '正在导出...' : '导出备份文件' }}</span>
            </button>
          </div>
        </div>

        <!-- Import Panel -->
        <div v-show="currentTab === 'import'" class="glass-panel rounded-2xl p-8 relative z-10">
          <div class="flex items-center gap-4 mb-6 relative z-10">
            <div class="w-14 h-14 rounded-xl bg-gradient-to-br from-cyber-accent2/20 to-purple-500/20 border border-cyber-accent2/30 flex items-center justify-center">
              <i class="fas fa-file-import text-2xl text-cyber-accent2"></i>
            </div>
            <div>
              <h2 class="text-lg font-semibold text-white">导入配置</h2>
              <p class="text-sm text-gray-500">从 .zip 备份文件恢复 SkillHub 配置</p>
            </div>
          </div>

          <!-- Drop Zone -->
          <div
            @click="selectImportFile"
            class="drop-zone rounded-2xl p-8 mb-6 text-center cursor-pointer relative z-10"
          >
            <div class="w-16 h-16 mx-auto mb-4 rounded-xl bg-cyber-panel flex items-center justify-center">
              <i class="fas fa-cloud-upload-alt text-2xl text-cyber-accent"></i>
            </div>
            <h3 class="text-lg font-semibold text-white mb-2">
              {{ importFile ? '已选择文件' : '选择备份文件' }}
            </h3>
            <p class="text-sm text-gray-500 mb-4">
              {{ importFile ? importFile.split('/').pop()?.split('\\').pop() : '支持 .zip 格式的 SkillHub 备份文件' }}
            </p>
            <button class="btn-secondary">
              <i class="fas fa-folder-open mr-2"></i>
              {{ importFile ? '重新选择' : '选择文件' }}
            </button>
          </div>

          <!-- Conflict Resolution -->
          <div class="mb-6 relative z-10">
            <label class="text-sm text-gray-400 block mb-3">冲突处理</label>
            <div class="space-y-3">
              <button
                @click="selectConflictMode('overwrite')"
                :class="['radio-btn w-full flex items-center gap-3 p-4 rounded-xl bg-cyber-dark border border-cyber-border cursor-pointer text-left', conflictMode === 'overwrite' ? 'selected' : '']"
              >
                <div class="radio-circle"></div>
                <div>
                  <p class="text-sm font-medium text-white">覆盖并更新</p>
                  <p class="text-xs text-gray-500 mt-1">用导入的配置替换现有配置</p>
                </div>
              </button>

              <button
                @click="selectConflictMode('skip')"
                :class="['radio-btn w-full flex items-center gap-3 p-4 rounded-xl bg-cyber-dark border border-cyber-border cursor-pointer text-left', conflictMode === 'skip' ? 'selected' : '']"
              >
                <div class="radio-circle"></div>
                <div>
                  <p class="text-sm font-medium text-white">跳过已有 Skill</p>
                  <p class="text-xs text-gray-500 mt-1">保留现有配置，仅导入新 Skill</p>
                </div>
              </button>
            </div>
          </div>

          <!-- Import Preview -->
          <div v-if="importPreview" class="mb-6 relative z-10">
            <div class="flex items-center justify-between mb-3">
              <label class="text-sm text-gray-400">导入预览</label>
              <span class="text-xs text-cyber-accent flex items-center gap-1">
                <i class="fas fa-check-circle"></i>
                文件有效
              </span>
            </div>
            <div class="p-4 rounded-xl bg-cyber-dark border border-cyber-border">
              <div class="flex items-center justify-between text-sm mb-2">
                <span class="text-gray-400">版本</span>
                <span class="text-white font-mono">{{ importPreview.version }}</span>
              </div>
              <div class="flex items-center justify-between text-sm mb-2">
                <span class="text-gray-400">导出日期</span>
                <span class="text-white font-mono">{{ importPreview.exportDate.split('T')[0] }}</span>
              </div>
              <div class="flex items-center justify-between text-sm mb-2">
                <span class="text-gray-400">包含 Skills</span>
                <span class="text-white font-mono">{{ importPreview.skillsCount }} 个</span>
              </div>
              <div class="flex items-center justify-between text-sm mb-2">
                <span class="text-gray-400">Git 缓存</span>
                <span class="text-white font-mono">{{ importPreview.gitCacheCount }} 个仓库</span>
              </div>
              <div class="flex items-center justify-between text-sm">
                <span class="text-gray-400">自定义工具</span>
                <span class="text-white font-mono">{{ importPreview.customToolsCount }} 个</span>
              </div>
            </div>
          </div>

          <!-- Import Button -->
          <div class="relative z-10">
            <button
              @click="startImport"
              :disabled="importing || !importFile"
              class="btn-primary w-full text-lg flex items-center justify-center gap-2"
            >
              <i :class="['fas', importing ? 'fa-spinner fa-spin' : 'fa-upload']"></i>
              <span>{{ importing ? '正在导入...' : '开始导入' }}</span>
            </button>
          </div>
        </div>

      </div>
    </main>
  </div>
</template>

<style scoped>
/* Glass Panel */
.glass-panel {
  position: relative;
  background: rgba(26, 26, 36, 0.6);
  backdrop-filter: blur(20px);
  border-radius: 16px;
}

.glass-panel::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  padding: 1px;
  background: linear-gradient(135deg,
    rgba(255, 255, 255, 0.12) 0%,
    rgba(255, 255, 255, 0.04) 50%,
    rgba(255, 255, 255, 0.12) 100%
  );
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  opacity: 0.8;
  transition: opacity 0.3s, background 0.3s;
  pointer-events: none;
}

.glass-panel:hover::before {
  opacity: 1;
  background: linear-gradient(135deg,
    rgba(0, 212, 170, 0.3) 0%,
    rgba(168, 85, 247, 0.2) 100%
  );
}

.glass-panel-no-hover:hover::before {
  background: linear-gradient(135deg,
    rgba(255, 255, 255, 0.12) 0%,
    rgba(255, 255, 255, 0.04) 50%,
    rgba(255, 255, 255, 0.12) 100%
  );
}

/* Header */
header.glass-panel {
  border-radius: 0 !important;
}

/* Icon Button */
.btn-icon {
  width: 36px;
  height: 36px;
  background: rgba(18, 18, 26, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #9ca3af;
  transition: all 0.2s ease;
}

.btn-icon:hover {
  border-color: rgba(0, 212, 170, 0.4);
  color: #00d4aa;
}

/* Primary Button */
.btn-primary {
  background: linear-gradient(135deg, #00d4aa, #a855f7);
  color: #0a0a0f;
  padding: 16px 32px;
  border-radius: 12px;
  font-weight: 600;
  border: none;
  cursor: pointer;
  box-shadow: 0 4px 16px rgba(0, 212, 170, 0.25);
  transition: all 0.3s ease;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 24px rgba(0, 212, 170, 0.4),
              0 0 30px rgba(0, 212, 170, 0.2);
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

/* Secondary Button */
.btn-secondary {
  background: rgba(18, 18, 26, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #9ca3af;
  padding: 12px 20px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-secondary:hover {
  border-color: rgba(0, 212, 170, 0.4);
  color: #00d4aa;
  background: rgba(0, 212, 170, 0.08);
}

/* Tab Button */
.tab-btn {
  transition: all 0.3s ease;
  border-bottom: 2px solid transparent;
}

.tab-btn:hover {
  color: #9ca3af;
}

.tab-btn.active {
  border-bottom-color: #00d4aa;
  color: #00d4aa;
}

/* Radio Button */
.radio-btn {
  transition: all 0.3s ease;
  position: relative;
}

.radio-btn::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  padding: 1px;
  background: linear-gradient(135deg,
    rgba(255, 255, 255, 0.08) 0%,
    rgba(255, 255, 255, 0.02) 100%
  );
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  pointer-events: none;
  transition: background 0.3s;
}

.radio-btn:hover::before {
  background: linear-gradient(135deg,
    rgba(0, 212, 170, 0.3) 0%,
    rgba(168, 85, 247, 0.2) 100%
  );
}

.radio-btn.selected {
  border-color: #00d4aa;
  background: rgba(0, 212, 170, 0.1);
}

.radio-btn.selected .radio-circle {
  border-color: #00d4aa;
  background: #00d4aa;
  box-shadow: 0 0 10px rgba(0, 212, 170, 0.3);
}

.radio-btn.selected .radio-circle::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 6px;
  height: 6px;
  background: #0a0a0f;
  border-radius: 50%;
}

.radio-circle {
  width: 18px;
  height: 18px;
  border: 2px solid #4b5563;
  border-radius: 50%;
  position: relative;
  transition: all 0.3s;
  flex-shrink: 0;
}

/* Drop Zone */
.drop-zone {
  transition: all 0.3s ease;
  border: 2px dashed rgba(255, 255, 255, 0.1);
}

.drop-zone:hover {
  border-color: rgba(0, 212, 170, 0.5);
  background: rgba(0, 212, 170, 0.05);
}

/* Gradient Text */
.gradient-text {
  background: linear-gradient(135deg, #00d4aa 0%, #a855f7 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
</style>
