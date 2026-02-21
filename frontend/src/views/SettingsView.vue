<script setup lang="ts">
import { ref, onMounted, computed, inject } from 'vue'
import { useRouter } from 'vue-router'
import { useSettingsStore } from '../stores/settings'
import { useSkillStore } from '../stores/skills'
import * as App from '../../wailsjs/go/backend/App'

const router = useRouter()
const settingsStore = useSettingsStore()
const skillStore = useSkillStore()

// 注入全局方法
const showNotification = inject<(message: string, type?: 'success' | 'error' | 'warning' | 'info', duration?: number) => void>('showNotification')
const showConfirm = inject<(options: any) => Promise<boolean | string>>('showConfirm')

const addingTool = ref(false)
const resetting = ref(false)
const checkingUpdate = ref(false)

// 更新设置相关
const autoUpdateCheck = ref(false)
const updateFrequency = ref<'startup' | 'daily' | 'weekly'>('daily')

// 自定义工具模态框
const showAddToolModal = ref(false)
const newToolName = ref('')
const newToolPath = ref('')

// 分类管理模态框
const showAddCategoryModal = ref(false)
const newCategoryName = ref('')
const addingCategory = ref(false)

// 工具颜色映射
const toolColors: Record<string, { from: string, to: string, text: string }> = {
  'claude-code': { from: 'from-orange-500/20', to: 'to-orange-600/20', text: 'text-orange-400' },
  'opencode': { from: 'from-blue-500/20', to: 'to-blue-600/20', text: 'text-blue-400' },
  'cursor': { from: 'from-purple-500/20', to: 'to-purple-600/20', text: 'text-purple-400' },
  'codebuddy': { from: 'from-green-500/20', to: 'to-green-600/20', text: 'text-green-400' },
  'trae': { from: 'from-pink-500/20', to: 'to-pink-600/20', text: 'text-pink-400' }
}

// 获取工具颜色配置
function getToolColor(toolId: string) { return toolColors[toolId] || { from: 'from-cyber-accent/20', to: 'to-cyber-accent2/20', text: 'text-cyber-accent' } }

// 分类颜色映射
const categoryColors: Record<string, { icon: string, colorClass: string }> = {
  '内容创作': { icon: 'fa-pen-fancy', colorClass: 'category-pink' },
  '认知增强': { icon: 'fa-lightbulb', colorClass: 'category-purple' },
  '开发辅助': { icon: 'fa-code', colorClass: 'category-orange' },
  '数据分析': { icon: 'fa-chart-bar', colorClass: 'category-cyan' },
  '教育学习': { icon: 'fa-graduation-cap', colorClass: 'category-green' },
  'AI/LLM': { icon: 'fa-brain', colorClass: 'category-blue' },
  '其他': { icon: 'fa-cube', colorClass: 'category-gray' }
}

// 获取分类颜色配置
function getCategoryColor(name: string) {
  return categoryColors[name] || { icon: 'fa-tag', colorClass: 'category-custom' }
}

// 工具信息类型
interface ToolInfo {
  id: string
  name: string
  skills_path: string
  is_installed: boolean
  is_enabled: boolean   // 工具是否启用（从配置文件读取）
}

// 自定义工具类型
interface CustomTool {
  id: string
  name: string
  skills_path: string
  enabled: boolean
  date_added: string
}

// 存储信息类型
interface StorageInfo {
  total_space: number
  used_space: number
  free_space: number
  skills_count: number
  skills_path: string
}

// 获取正确的路径分隔符
function getPathSeparator(): string {
  // 在 Windows 上路径使用反斜杠，其他系统使用正斜杠
  return navigator.platform.toLowerCase().includes('win') ? '\\' : '/'
}

const detectedTools = ref<ToolInfo[]>([])
const customTools = ref<CustomTool[]>([])
const storageInfo = ref<StorageInfo | null>(null)
const loadingStorage = ref(false)

// 更新频率选项
const updateFrequencies = [
  { value: 'startup', label: '每次启动' },
  { value: 'daily', label: '每天' },
  { value: 'weekly', label: '每周' }
]

// 计算存储使用率百分比
const usagePct = computed(() => {
  if (!storageInfo.value || storageInfo.value.total_space === 0) return 0
  return Math.round((storageInfo.value.used_space / storageInfo.value.total_space) * 100)
})

// 计算存储状态
const storageStatus = computed(() => {
  if (!storageInfo.value) return { text: '加载中...', class: 'text-gray-400' }

  const pct = usagePct.value
  if (pct < 60) {
    return { text: '✓ 存储空间充足', class: 'text-green-400' }
  } else if (pct < 80) {
    return { text: '⚠ 存储空间尚可', class: 'text-yellow-400' }
  } else {
    return { text: '✗ 存储空间不足', class: 'text-red-400' }
  }
})

// 加载数据
onMounted(async () => {
  await settingsStore.loadSettings() // 先加载设置（包括路径）
  await loadTools()
  await loadStorageInfo()
  await loadUpdateSettings()
  await skillStore.loadCategories() // 加载分类数据
})

async function loadTools() {
  try {
    // 加载检测到的工具
    detectedTools.value = await App.GetDetectedTools()
    // 加载自定义工具
    const custom = await App.GetCustomTools()
    customTools.value = custom || []
  } catch (e) {
    console.error('Failed to load tools:', e)
  }
}

async function loadStorageInfo() {
  loadingStorage.value = true
  try {
    storageInfo.value = await App.GetStorageInfo()
  } catch (e) {
    console.error('Failed to load storage info:', e)
  } finally {
    loadingStorage.value = false
  }
}

async function loadUpdateSettings() {
  try {
    const freq = await App.GetUpdateFrequency()
    updateFrequency.value = (freq as 'startup' | 'daily' | 'weekly') || 'daily'
    autoUpdateCheck.value = freq !== 'disabled'
  } catch (e) {
    console.error('Failed to load update settings:', e)
  }
}

async function toggleAutoUpdate() {
  try {
    if (autoUpdateCheck.value) {
      await App.SetUpdateFrequency(updateFrequency.value)
    } else {
      await App.SetUpdateFrequency('disabled')
    }
  } catch (e) {
    console.error('Failed to toggle auto update:', e)
  }
}

async function setUpdateFrequency(freq: 'startup' | 'daily' | 'weekly') {
  updateFrequency.value = freq
  try {
    await App.SetUpdateFrequency(freq)
  } catch (e) {
    console.error('Failed to set update frequency:', e)
  }
}

function goBack() {
  router.push('/')
}

// 切换工具启用状态
async function toggleTool(toolId: string) {
  try {
    // 检查工具是否已安装
    const tool = detectedTools.value.find(t => t.id === toolId)
    if (!tool || !tool.is_installed) {
      showNotification?.('未检测的工具无法启用，请先安装该 AI 工具', 'warning')
      return
    }

    await App.ToggleToolEnabled(toolId)
    // 本地更新状态，不重新加载避免顺序跳动
    tool.is_enabled = !tool.is_enabled
  } catch (e) {
    console.error('Toggle tool failed:', e)
    showNotification?.('切换工具状态失败: ' + e, 'error')
  }
}

// 打开添加工具模态框
function openAddToolModal() {
  newToolName.value = ''
  newToolPath.value = ''
  showAddToolModal.value = true
}

// 关闭模态框
function closeAddToolModal() {
  showAddToolModal.value = false
}

// 浏览文件夹
async function browseForToolPath() {
  try {
    const path = await App.SelectFolder()
    if (path) {
      newToolPath.value = path
      // 确保路径以正确的分隔符结尾
      const separator = getPathSeparator()
      if (!path.endsWith(separator)) {
        newToolPath.value = path + separator
      }
    }
  } catch (e) {
    console.error('选择文件夹失败:', e)
  }
}

// 确认添加工具
async function confirmAddTool() {
  if (!newToolName.value.trim() || !newToolPath.value.trim()) {
    showNotification?.('请填写工具名称和路径', 'warning')
    return
  }

  addingTool.value = true
  try {
    const tool: CustomTool = {
      id: 'custom-' + Date.now(),
      name: newToolName.value.trim(),
      skills_path: newToolPath.value.trim(),
      enabled: true,
      date_added: new Date().toISOString().split('T')[0]
    }

    await App.AddCustomTool(tool)
    closeAddToolModal()
    await loadTools() // 重新加载
    showNotification?.('工具添加成功', 'success')
  } catch (e) {
    showNotification?.('添加工具失败: ' + e, 'error')
  } finally {
    addingTool.value = false
  }
}

// 删除自定义工具
async function removeCustomTool(id: string) {
  const confirmed = showConfirm ? await showConfirm({
    title: '删除自定义工具',
    message: '确定要删除这个自定义工具吗？\n此操作无法撤销。',
    type: 'danger',
    confirmText: '删除',
    cancelText: '取消'
  }) : confirm('确定要删除这个自定义工具吗？')

  if (!confirmed) {
    return
  }

  try {
    await App.RemoveCustomTool(id)
    await loadTools() // 重新加载
    showNotification?.('工具删除成功', 'success')
  } catch (e) {
    console.error('Remove custom tool failed:', e)
    showNotification?.('删除工具失败: ' + e, 'error')
  }
}

// 修改存储路径（支持数据迁移）
async function changeStoragePath() {
  try {
    const path = await App.SelectFolder()
    if (!path || path === '') {
      return
    }

    // 获取迁移信息
    const migrationInfo = await App.GetMigrationInfo(path)

    // 如果有旧数据，弹出确认对话框
    if (migrationInfo.has_old_data) {
      const confirmed = showConfirm ? await showConfirm({
        title: '迁移数据到新路径',
        message: `检测到原路径有 ${migrationInfo.skills_count} 个 Skills。是否迁移到新路径？`,
        details: {
          '原路径': migrationInfo.old_path,
          '新路径': migrationInfo.new_path,
          'Skills 数量': migrationInfo.skills_count.toString()
        },
        type: 'info',
        confirmText: '迁移数据',
        cancelText: '取消'
      }) : confirm(`检测到原路径有 ${migrationInfo.skills_count} 个 Skills。是否迁移到新路径？\n\n确定=迁移，取消=不迁移`)

      if (!confirmed) {
        return
      }

      // 显示进度提示
      showNotification?.('正在迁移数据...', 'info')

      // 执行迁移
      await App.SetSkillHubPathWithMigration(path)
      settingsStore.skillhubPath = path

      // 迁移成功后，提醒用户自行检查和删除原路径
      showNotification?.('数据已迁移至新路径。原路径数据保留，您可以自行检查后手动删除：' + migrationInfo.old_path, 'success')
    } else {
      // 无旧数据，直接切换
      await App.SetSkillHubPath(path)
      settingsStore.skillhubPath = path
      showNotification?.('存储路径已更新', 'success')
    }

    // 重新加载存储信息
    await loadStorageInfo()
  } catch (e) {
    console.error('Change storage path failed:', e)
    showNotification?.('修改存储路径失败: ' + e, 'error')
  }
}

// 导出数据
// 重置所有数据
async function resetAllData() {
  const confirmed = showConfirm ? await showConfirm({
    title: '重置数据确认',
    message: '确定要重置所有数据吗？\n\n此操作将：\n• 删除所有已安装的 Skills\n• 清空所有元数据\n• 清除操作日志\n\n此操作无法撤销！',
    type: 'danger',
    confirmText: '重置',
    cancelText: '取消'
  }) : confirm('确定要重置所有数据吗？此操作无法撤销！')

  if (!confirmed) {
    return
  }

  resetting.value = true
  try {
    await App.ResetAllData()
    showNotification?.('数据已重置', 'success')
    location.reload()
  } catch (e) {
    showNotification?.('重置失败: ' + e, 'error')
  } finally {
    resetting.value = false
  }
}

// 检查更新
async function checkForUpdates() {
  checkingUpdate.value = true
  try {
    const updateInfo = await App.CheckForUpdates()
    if (updateInfo.update_count > 0) {
      const skills = updateInfo.skills_with_update.join(', ')
      showNotification?.(`${updateInfo.update_count} 个 Skill 有更新: ${skills}`, 'info')
    } else {
      showNotification?.('所有 Skills 都是最新版本', 'success')
    }
  } catch (e) {
    showNotification?.('检查更新失败: ' + e, 'error')
  } finally {
    checkingUpdate.value = false
  }
}

// 打开添加分类模态框
function openAddCategoryModal() {
  newCategoryName.value = ''
  showAddCategoryModal.value = true
}

// 关闭分类模态框
function closeAddCategoryModal() {
  showAddCategoryModal.value = false
}

// 确认添加分类
async function confirmAddCategory() {
  if (!newCategoryName.value.trim()) {
    showNotification?.('请输入分类名称', 'warning')
    return
  }

  addingCategory.value = true
  try {
    await skillStore.addCategory(newCategoryName.value.trim())
    closeAddCategoryModal()
    showNotification?.('分类添加成功', 'success')
  } catch (e) {
    showNotification?.('添加分类失败: ' + e, 'error')
  } finally {
    addingCategory.value = false
  }
}

// 删除分类
async function deleteCategory(name: string) {
  const confirmed = showConfirm ? await showConfirm({
    title: '删除分类',
    message: `确定要删除分类「${name}」吗？\n\n使用该分类的 Skills 将被自动改为「其他」分类。`,
    type: 'danger',
    confirmText: '删除',
    cancelText: '取消'
  }) : confirm(`确定要删除分类「${name}」吗？`)

  if (!confirmed) {
    return
  }

  try {
    const affectedSkills = await skillStore.deleteCategory(name)
    if (affectedSkills && affectedSkills.length > 0) {
      showNotification?.(`分类已删除，${affectedSkills.length} 个 Skill 已改为"其他"分类`, 'success')
    } else {
      showNotification?.('分类删除成功', 'success')
    }
  } catch (e) {
    showNotification?.('删除分类失败: ' + e, 'error')
  }
}
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
              <span class="gradient-text">设置</span>
            </h1>
            <p class="text-xs text-gray-500">配置 SkillHub 外观和行为</p>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="flex-1 overflow-y-auto p-6">
      <div class="w-full max-w-2xl mx-auto space-y-5">
        <!-- Storage Settings -->
        <div class="glass-panel setting-card rounded-2xl p-6 relative z-10">
          <h2 class="text-lg font-semibold text-white mb-5 flex items-center gap-3">
            <i class="fas fa-hdd text-cyber-accent"></i>
            存储
          </h2>

          <div class="space-y-5">
            <!-- Storage Path -->
            <div>
              <label class="text-sm text-gray-400 block mb-2">SkillHub 路径</label>
              <div class="flex gap-3">
                <code class="flex-1 px-4 py-3 rounded-xl bg-cyber-dark border border-cyber-border text-sm font-mono text-gray-300">
                  {{ settingsStore.skillhubPath }}
                </code>
                <button @click="changeStoragePath" class="btn-secondary">
                  更改
                </button>
              </div>
            </div>

            <!-- Storage Space Card -->
            <div v-if="loadingStorage" class="flex items-center justify-center py-6">
              <i class="fas fa-spinner fa-spin text-gray-400 mr-2"></i>
              <span class="text-sm text-gray-500">加载中...</span>
            </div>
            <div v-else-if="storageInfo" class="p-4 rounded-xl bg-cyber-dark/50 border border-cyber-border">
              <div class="flex items-center justify-between text-sm">
                <span class="text-gray-400">存储空间</span>
                <span class="font-mono text-white">{{ storageInfo.used_space }} MB / {{ storageInfo.total_space }} MB</span>
              </div>
              <div class="mt-3 h-2 bg-cyber-dark rounded-full overflow-hidden">
                <div
                  class="h-full bg-gradient-to-r from-cyber-accent to-cyber-accent2 rounded-full transition-all"
                  :style="{ width: usagePct + '%' }"
                ></div>
              </div>
              <p class="mt-2 text-xs text-gray-500">磁盘空间充足</p>
            </div>
          </div>
        </div>

        <!-- Appearance Settings -->
        <div class="glass-panel rounded-2xl p-6 border border-cyber-border">
          <h2 class="text-lg font-semibold text-white mb-4 flex items-center gap-3">
            <i class="fas fa-palette text-xl text-cyber-accent"></i>
            外观
          </h2>
          <div class="space-y-4">
            <label class="text-sm text-gray-400 block mb-3">主题模式</label>
            <div class="grid grid-cols-3 gap-3">
              <!-- System Theme -->
              <button
                @click="settingsStore.setTheme('system')"
                :class="[
                  'radio-btn flex items-center gap-3 p-4 rounded-lg bg-cyber-dark border border-cyber-border cursor-pointer transition-all',
                  settingsStore.theme === 'system' ? 'selected' : ''
                ]"
              >
                <div class="radio-circle"></div>
                <div class="text-left">
                  <p class="text-sm font-medium text-white">跟随系统</p>
                  <p class="text-xs text-gray-500">自动切换</p>
                </div>
              </button>

              <!-- Dark Theme -->
              <button
                @click="settingsStore.setTheme('dark')"
                :class="[
                  'radio-btn flex items-center gap-3 p-4 rounded-lg bg-cyber-dark border border-cyber-border cursor-pointer transition-all',
                  settingsStore.theme === 'dark' ? 'selected' : ''
                ]"
              >
                <div class="radio-circle"></div>
                <div class="text-left">
                  <p class="text-sm font-medium text-white">暗黑模式</p>
                  <p class="text-xs text-gray-500">深色主题</p>
                </div>
              </button>

              <!-- Light Theme -->
              <button
                @click="settingsStore.setTheme('light')"
                :class="[
                  'radio-btn flex items-center gap-3 p-4 rounded-lg bg-cyber-dark border border-cyber-border cursor-pointer transition-all',
                  settingsStore.theme === 'light' ? 'selected' : ''
                ]"
              >
                <div class="radio-circle"></div>
                <div class="text-left">
                  <p class="text-sm font-medium text-white">亮色模式</p>
                  <p class="text-xs text-gray-500">浅色主题</p>
                </div>
              </button>
            </div>
          </div>
        </div>

        <!-- Category Management -->
        <div class="glass-panel rounded-2xl p-6 border border-cyber-border">
          <div class="flex items-center justify-between mb-5">
            <h2 class="text-lg font-semibold text-white flex items-center gap-3">
              <i class="fas fa-tags text-xl text-cyber-accent"></i>
              分类管理
            </h2>
            <button @click="openAddCategoryModal" class="btn-secondary px-4 py-2 text-sm">
              <i class="fas fa-plus mr-2"></i>
              添加分类
            </button>
          </div>

          <!-- Preset Categories -->
          <div class="mb-5">
            <p class="text-xs text-gray-500 mb-3">预设分类（不可删除）</p>
            <div class="flex flex-wrap gap-2">
              <span
                v-for="cat in skillStore.allCategoriesWithInfo.filter(c => c.is_preset)"
                :key="cat.name"
                :class="['category-tag', getCategoryColor(cat.name).colorClass]"
              >
                <i :class="['fas', getCategoryColor(cat.name).icon, 'mr-1.5']"></i>
                {{ cat.name }}
              </span>
            </div>
          </div>

          <!-- Custom Categories -->
          <div>
            <p class="text-xs text-gray-500 mb-3">自定义分类</p>
            <div v-if="skillStore.customCategories.length === 0" class="text-sm text-gray-500 italic">
              暂无自定义分类，点击上方按钮添加
            </div>
            <div v-else class="flex flex-wrap gap-2">
              <span
                v-for="name in skillStore.customCategories"
                :key="name"
                class="category-tag category-custom"
              >
                <i class="fas fa-tag mr-1.5"></i>
                {{ name }}
                <button
                  @click.stop="deleteCategory(name)"
                  class="category-delete-btn"
                  title="删除分类"
                >
                  <i class="fas fa-times text-white text-[10px]"></i>
                </button>
              </span>
            </div>
          </div>
        </div>

        <!-- Auto Update Settings -->
        <div class="glass-panel rounded-2xl p-6 border border-cyber-border">
          <h2 class="text-lg font-semibold text-white mb-4 flex items-center gap-3">
            <i class="fas fa-sync text-xl text-cyber-accent2"></i>
            检查SKill更新
          </h2>
          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <label class="text-sm text-gray-400 block mb-3">更新频率</label>
              <button
                @click="autoUpdateCheck = !autoUpdateCheck; toggleAutoUpdate()"
                :class="[
                  'toggle-switch w-12 h-6 rounded-full relative transition-all',
                  autoUpdateCheck ? 'active' : 'bg-cyber-dark border border-cyber-border'
                ]"
                title="启用/禁用自动检查更新"
              ></button>
            </div>
            <div v-show="autoUpdateCheck" class="flex gap-3">
              <button
                v-for="freq in updateFrequencies"
                :key="freq.value"
                @click="setUpdateFrequency(freq.value as 'startup' | 'daily' | 'weekly')"
                :class="[
                  'btn-frequency',
                  updateFrequency === freq.value ? 'active' : ''
                ]"
              >
                {{ freq.label }}
              </button>
            </div>
          </div>
        </div>

        <!-- AI Tools Management -->
        <div class="glass-panel rounded-2xl p-6 border border-cyber-border">
          <h2 class="text-lg font-semibold text-white flex items-center gap-3">
            <i class="fas fa-robot text-xl text-cyber-accent"></i>
            AI 工具
            <span class="text-xs text-gray-500 bg-cyber-dark px-2 py-1 rounded-lg ml-2">
              {{ detectedTools.length }} 个
            </span>
          </h2>
          <div v-if="detectedTools.length > 0" class="space-y-3 mt-5">
            <div
              v-for="tool in detectedTools"
              :key="tool.id"
              @click="toggleTool(tool.id)"
              :class="[
                'tool-card p-4 relative',
                tool.is_enabled ? 'selected' : '',
                !tool.is_installed ? 'disabled' : ''
              ]"
            >
              <div class="flex items-center justify-between relative z-10">
                <!-- 左侧：工具图标和信息 -->
                <div class="flex items-center gap-4">
                  <div class="w-10 h-10 rounded-lg bg-gradient-to-br flex items-center justify-center"
                       :class="[getToolColor(tool.id).from, getToolColor(tool.id).to]">
                    <span class="text-sm font-bold" :class="getToolColor(tool.id).text">
                      {{ tool.name.charAt(0) }}
                    </span>
                  </div>
                  <div>
                    <h4 class="font-medium text-white">{{ tool.name }}</h4>
                    <p class="text-xs text-gray-500">{{ tool.skills_path }}</p>
                  </div>
                </div>
                <!-- 右侧：标签和勾选图标 -->
                <div class="flex items-center gap-3">
                  <span class="tag" :class="{ muted: !tool.is_installed }">
                    {{ tool.is_installed ? '已检测' : '未检测' }}
                  </span>
                  <div class="check-icon w-6 h-6 rounded-full bg-cyber-accent flex items-center justify-center">
                    <i class="fas fa-check text-cyber-dark text-xs"></i>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Custom Tools Management -->
        <div class="glass-panel rounded-2xl p-6 border border-cyber-border">
          <div class="flex items-center justify-between mb-5">
            <h2 class="text-lg font-semibold text-white flex items-center gap-3">
              <i class="fas fa-toolbox text-xl text-cyber-accent"></i>
              自定义工具
            </h2>
            <button @click="openAddToolModal" class="btn-secondary px-4 py-2 text-sm">
              <i class="fas fa-plus mr-2"></i>
              添加
            </button>
          </div>
          <div class="space-y-3 mt-5">
            <div
              v-for="tool in customTools"
              :key="tool.id"
              class="tool-card p-4 relative"
            >
              <div class="flex items-center justify-between relative z-10">
                <!-- 左侧：工具图标和信息 -->
                <div class="flex items-center gap-4">
                  <div class="w-10 h-10 rounded-lg bg-gradient-to-br flex items-center justify-center from-cyber-accent/20 to-cyber-accent2/20">
                    <span class="text-sm font-bold text-cyber-accent">
                      {{ tool.name.charAt(0) }}
                    </span>
                  </div>
                  <div>
                    <h4 class="font-medium text-white">{{ tool.name }}</h4>
                    <p class="text-xs text-gray-500">{{ tool.skills_path }}</p>
                  </div>
                </div>
                <!-- 右侧：删除按钮 -->
                <button
                  @click="removeCustomTool(tool.id)"
                  class="btn-frequency text-sm flex-shrink-0"
                >
                  <i class="fas fa-trash-alt"></i>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Data & History -->
        <div class="glass-panel rounded-2xl p-6 border border-cyber-border">
          <h2 class="text-lg font-semibold text-white mb-5 flex items-center gap-3">
            <i class="fas fa-database text-xl text-cyber-accent2"></i>
            数据管理
          </h2>
          <div class="grid grid-cols-2 gap-3">
            <!-- History -->
            <button @click="router.push('/history')" class="data-card flex flex-col items-center text-center p-4 rounded-lg bg-cyber-dark border border-cyber-border hover:border-cyber-accent/50 transition-all group">
              <i class="fas fa-history text-gray-500 group-hover:text-cyber-accent transition-colors mb-2"></i>
              <h4 class="text-sm font-medium text-white mb-1">操作历史</h4>
              <p class="text-xs text-gray-500">查看最近 10 天操作</p>
            </button>
            <!-- Export -->
            <button @click="router.push('/import-export')" class="data-card flex flex-col items-center text-center p-4 rounded-lg bg-cyber-dark border border-cyber-border hover:border-cyber-accent/50 transition-all group">
              <i class="fas fa-exchange-alt text-gray-500 group-hover:text-cyber-accent transition-colors mb-2"></i>
              <h4 class="text-sm font-medium text-white mb-1">导入/导出</h4>
              <p class="text-xs text-gray-500">备份与恢复</p>
            </button>
            <!-- Reset Data -->
            <button
              @click="resetAllData"
              :disabled="resetting"
              class="data-card flex flex-col items-center text-center p-4 rounded-lg bg-cyber-dark border border-cyber-border hover:border-red-400/50 hover:bg-red-500/5 transition-all group disabled:opacity-50 disabled:cursor-not-allowed"
              >
              <i :class="['fas', resetting ? 'fa-spinner fa-spin' : 'fa-trash-alt', 'text-gray-500 group-hover:text-red-400 transition-colors mb-2']"></i>
              <h4 class="text-sm font-medium text-white mb-1">重置数据</h4>
              <p class="text-xs text-gray-500">清除所有配置</p>
            </button>
            <!-- Check Updates -->
            <button
              @click="checkForUpdates"
              :disabled="checkingUpdate"
              class="data-card flex flex-col items-center text-center p-4 rounded-lg bg-cyber-dark border border-cyber-border hover:border-cyber-accent/50 transition-all group disabled:opacity-50 disabled:cursor-not-allowed"
              >
              <i :class="['fas', checkingUpdate ? 'fa-spinner fa-spin' : 'fa-cloud-download-alt', 'text-gray-500 group-hover:text-cyber-accent transition-colors mb-2']"></i>
              <h4 class="text-sm font-medium text-white mb-1">检查更新</h4>
              <p class="text-xs text-gray-500">当前 v1.0.0</p>
            </button>
          </div>
        </div>

        <!-- Footer Info -->
        <div class="text-center text-sm text-gray-600 space-y-1">
          <p>SkillHub v1.0.0</p>
          <p>开源软件 · MIT License</p>
        </div>
      </div>
    </main>

    <!-- Add Tool Modal -->
    <div v-if="showAddToolModal" class="modal-overlay" @click.self="closeAddToolModal">
      <div class="modal-content glass-panel">
        <!-- Modal Header -->
        <div class="modal-header">
          <h3 class="text-lg font-semibold text-white">添加自定义工具</h3>
          <button @click="closeAddToolModal" class="modal-close">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <!-- Modal Body -->
        <div class="modal-body space-y-4">
          <!-- Tool Name -->
          <div>
            <label class="block text-sm font-medium text-gray-400 mb-2">工具名称</label>
            <input
              v-model="newToolName"
              type="text"
              placeholder="例如: MyCustomTool"
              class="modal-input w-full py-3 px-4 text-sm"
            >
          </div>

          <!-- Skills Path -->
          <div>
            <label class="block text-sm font-medium text-gray-400 mb-2">Skills 路径</label>
            <div class="flex gap-3">
              <div class="flex-1 relative">
                <i class="fas fa-folder absolute left-4 top-1/2 -translate-y-1/2 text-gray-500"></i>
                <input
                  v-model="newToolPath"
                  type="text"
                  placeholder="~/.mytool/skills/"
                  class="modal-input w-full py-3 pl-11 pr-4 text-sm font-mono"
                >
              </div>
              <button
                @click="browseForToolPath"
                class="modal-btn px-3 py-2 bg-cyber-panel border border-cyber-border text-sm"
                >
                <i class="fas fa-folder-open"></i>
                浏览
              </button>
            </div>
          </div>

          <div class="flex justify-end gap-3">
            <button
              @click="confirmAddTool"
              :disabled="addingTool"
              class="modal-btn px-4 py-3 bg-cyber-accent/10 border border-cyber-accent/30 text-sm"
              >
              <i v-if="addingTool" class="fas fa-spinner fa-spin mr-2"></i>
              确认添加
              </button>
            <button
              @click="closeAddToolModal"
              class="modal-btn px-4 py-3 bg-cyber-panel border border-cyber-border text-sm"
              >
              取消
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Add Category Modal -->
    <div v-if="showAddCategoryModal" class="modal-overlay" @click.self="closeAddCategoryModal">
      <div class="modal-content glass-panel">
        <!-- Modal Header -->
        <div class="modal-header">
          <h3 class="text-lg font-semibold text-white">添加自定义分类</h3>
          <button @click="closeAddCategoryModal" class="modal-close">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <!-- Modal Body -->
        <div class="modal-body space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-400 mb-2">分类名称</label>
            <input
              v-model="newCategoryName"
              type="text"
              placeholder="例如: 办公效率"
              class="modal-input w-full py-3 px-4 text-sm"
              @keyup.enter="confirmAddCategory"
            >
          </div>

          <div class="flex justify-end gap-3">
            <button
              @click="confirmAddCategory"
              :disabled="addingCategory"
              class="modal-btn px-4 py-3 bg-cyber-accent/10 border border-cyber-accent/30 text-sm"
              >
              <i v-if="addingCategory" class="fas fa-spinner fa-spin mr-2"></i>
              确认添加
              </button>
            <button
              @click="closeAddCategoryModal"
              class="modal-btn px-4 py-3 bg-cyber-panel border border-cyber-border text-sm"
              >
              取消
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Glass Panel 卡片悬停效果 - 与原型一致 */
.glass-panel {
  position: relative;
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

/* Header */
header {
  background: rgba(26, 26, 36, 0.6, 0.8);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

/* 修复 Header 圆角 - 覆盖全局 glass-panel 的圆角定义 */
header.glass-panel {
  border-radius: 0 !important;
}

.glass-panel-no-hover:hover::before {
  background: linear-gradient(135deg,
    rgba(255, 255, 255, 0.12) 0%,
    rgba(255, 255, 255, 0.04) 50%,
    rgba(255, 255, 255, 0.12) 100%
  );
}

/* Buttons */
.btn-primary {
  background: linear-gradient(135deg, #00d4aa, #a855f7);
  color: #0a0a0f;
  padding: 12px 32px;
  border-radius: 12px;
  font-weight: 600;
  border: none;
  cursor: pointer;
  box-shadow: 0 4px 16px rgba(0, 212, 170, 0.25);
  transition: all 0.3s ease;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 24px rgba(0, 212, 170, 0.4),
              0 0 30px rgba(0, 212, 170, 0.2);
}

.btn-primary:active {
  transform: translateY(0) scale(0.98);
}

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

.btn-ghost {
  background: rgba(18, 18, 26, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #9ca3af;
  padding: 10px 16px;
  border-radius: 12px;
  cursor: pointer;
  button: auto;
  transition: all 0.2s ease;
}

.btn-ghost:hover {
  background: rgba(18, 18, 26, 0.8);
  border-color: rgba(0, 212, 170, 0.4);
  color: #d1d5db;
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
  title: "启用/禁用自动检查更新";
  transition: all 0.2s ease;
}

.btn-icon:hover {
  border-color: rgba(0, 212, 170, 0.4);
  color: #00d4aa;
}

/* Toggle Switch */
.toggle-switch {
  position: relative;
  width: 48px;
  height: 26px;
  background: #1e1e2e;
  border-radius: 13px;
  cursor: pointer;
  transition: all 0.3s ease;
  border: 2px solid transparent;
}

.toggle-switch:hover {
  border-color: rgba(0, 212, 170, 0.3);
}

.toggle-switch.active {
  background: linear-gradient(135deg, #00d4aa, #00b894);
  box-shadow: 0 0 20px rgba(0, 212, 170, 0.3);
}

.toggle-switch::after {
  content: '';
  position: absolute;
  top: 3px;
  left: 3px;
  width: 20px;
  height: 20px;
  background: white;
  border-radius: 50%;
  transition: transform 0.3s;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.toggle-switch.active::after {
  transform: translateX(22px);
}

/* Frequency Button */
.btn-frequency {
  padding: 12px 16px;
  border-radius: 12px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s ease;
  background: rgba(18, 18, 26, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #9ca3af;
}

.btn-frequency:hover {
  border-color: rgba(0, 212, 170, 0.4);
  color: #00d4aa;
}

.btn-frequency.active {
  background: linear-gradient(135deg, #00d4aa, #a855f7);
  color: #0a0a0f;
  font-weight: 600;
  border: none;
}

.btn-frequency.active:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0, 212, 170, 0.25);
}

/* Radio Button - 主题选择器 */
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
  transform: scale(1);
}

.radio-circle {
  width: 20px;
  height: 20px;
  border: 2px solid #4b5563;
  border-radius: 50%;
  position: relative;
  transition: all 0.3s;
  flex-shrink: 0;
}

.radio-circle::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%) scale(0);
  width: 8px;
  height: 8px;
  background: #0a0a0f;
  border-radius: 50%;
  transition: transform 0.3s;
}

/* Data Card */
.data-card:hover {
  transform: translateY(-2px);
}

/* Tool Card - 工具卡片样式 */
.tool-card {
  position: relative;
  background: rgba(26, 26, 36, 0.6);
  backdrop-filter: blur(12px);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.tool-card::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  padding: 1px;
  background: linear-gradient(135deg,
    rgba(255, 255, 255, 0.1) 0%,
    rgba(255, 255, 255, 0.03) 50%,
    rgba(255, 255, 255, 0.1) 100%
  );
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  opacity: 0.7;
  transition: opacity 0.3s, background 0.3s;
  pointer-events: none;
}

.tool-card:hover {
  transform: translateY(-2px);
}

.tool-card:hover::before {
  opacity: 1;
  background: linear-gradient(135deg,
    rgba(0, 212, 170, 0.35) 0%,
    rgba(168, 85, 247, 0.25) 100%
  );
}

.tool-card.selected::before {
  opacity: 1;
  background: linear-gradient(135deg,
    rgba(0, 212, 170, 0.5) 0%,
    rgba(168, 85, 247, 0.4) 100%
  );
}

.tool-card.selected {
  background: rgba(0, 212, 170, 0.08);
}

.tool-card.selected .check-icon {
  opacity: 1;
  transform: scale(1);
}

.check-icon {
  opacity: 0;
  transform: scale(0);
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

/* Tag 标签样式 */
.tag {
  display: inline-flex;
  align-items: center;
  padding: 4px 8px;
  background: rgba(0, 212, 170, 0.1);
  border: 1px solid rgba(0, 212, 170, 0.25);
  border-radius: 6px;
  font-size: 12px;
  color: #00d4aa;
}

.tag.muted {
  background: rgba(107, 114, 128, 0.15);
  border-color: rgba(107, 114, 128, 0.3);
  color: #9ca3af;
}

/* 未安装工具禁用状态 */
.tool-card.disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.tool-card.disabled:hover {
  transform: none;
}

.tool-card.disabled:hover::before {
  background: linear-gradient(135deg,
    rgba(255, 255, 255, 0.1) 0%,
    rgba(255, 255, 255, 0.03) 50%,
    rgba(255, 255, 255, 0.1) 100%
  );
}

/* Modal 模态框样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}

.modal-content {
  width: 100%;
  max-width: 480px;
  max-height: 90vh;
  overflow-y: auto;
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #fff;
}

.modal-close {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  color: #9ca3af;
  cursor: pointer;
  transition: all 0.2s ease;
}

.modal-close:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
  border-color: rgba(255, 255, 255, 0.2);
}

.modal-body {
  padding: 24px;
}

.modal-input {
  background: rgba(18, 18, 26, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  color: #fff;
  transition: all 0.2s ease;
}

.modal-input:focus {
  outline: none;
  border-color: rgba(0, 212, 170, 0.5);
  box-shadow: 0 0 0 3px rgba(0, 212, 170, 0.1);
}

.modal-input::placeholder {
  color: #6b7280;
}

/* Modal 按钮样式 */
.modal-btn {
  border-radius: 10px;
  transition: all 0.2s ease;
}

.modal-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
  border-color: rgba(255, 255, 255, 0.2);
}

/* Category Tag 分类标签样式 */
.category-tag {
  display: inline-flex;
  align-items: center;
  padding: 6px 12px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s ease;
}

/* 预设分类颜色 */
.category-pink {
  background: rgba(236, 72, 153, 0.15);
  border: 1px solid rgba(236, 72, 153, 0.3);
  color: #ec4899;
}

.category-purple {
  background: rgba(168, 85, 247, 0.15);
  border: 1px solid rgba(168, 85, 247, 0.3);
  color: #a855f7;
}

.category-orange {
  background: rgba(249, 115, 22, 0.15);
  border: 1px solid rgba(249, 115, 22, 0.3);
  color: #f97316;
}

.category-cyan {
  background: rgba(6, 182, 212, 0.15);
  border: 1px solid rgba(6, 182, 212, 0.3);
  color: #06b6d4;
}

.category-green {
  background: rgba(34, 197, 94, 0.15);
  border: 1px solid rgba(34, 197, 94, 0.3);
  color: #22c55e;
}

.category-blue {
  background: rgba(59, 130, 246, 0.15);
  border: 1px solid rgba(59, 130, 246, 0.3);
  color: #3b82f6;
}

.category-gray {
  background: rgba(107, 114, 128, 0.15);
  border: 1px solid rgba(107, 114, 128, 0.3);
  color: #9ca3af;
}

/* 自定义分类颜色 */
.category-custom {
  background: rgba(0, 212, 170, 0.15);
  border: 1px solid rgba(0, 212, 170, 0.3);
  color: #00d4aa;
}

.category-custom:hover {
  background: rgba(0, 212, 170, 0.25);
}

/* 自定义分类删除按钮 - 使用原生 CSS 替代 group-hover */
.category-tag:has(.category-delete-btn) {
  position: relative;
}

.category-delete-btn {
  position: absolute;
  top: -6px;
  right: -6px;
  width: 18px;
  height: 18px;
  background: #ef4444;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s ease;
  cursor: pointer;
  border: 1px solid rgba(255, 255, 255, 0.2);
  z-index: 10;
}

.category-tag:hover .category-delete-btn {
  opacity: 1;
}

.category-delete-btn:hover {
  background: #dc2626;
  transform: scale(1.1);
}
</style>