<script setup lang="ts">
import { onMounted, ref, computed, inject } from 'vue'
import { useRouter } from 'vue-router'
import { useSkillStore } from '../stores/skills'
import { useSettingsStore } from '../stores/settings'
import * as App from '../../wailsjs/go/backend/App'

const router = useRouter()
const skillStore = useSkillStore()
const settingsStore = useSettingsStore()

// 注入全局方法
const showConfirm = inject<(options: any) => Promise<boolean>>('showConfirm')
const showNotification = inject<(message: string, type: string) => void>('showNotification')

const expandedSkill = ref<string | null>(null)
const updatingSkill = ref<string | null>(null)
const batchUpdatingProgress = ref(0) // 批量更新进度

// 编辑状态 - 使用 reactive 对象存储
const editingData = ref<Record<string, {
  category: string
  tags: string[]
  notes: string
  newTag: string
}>>({})

// 检查数据是否有变化
function hasChanges(skillId: string): boolean {
  const skill = skillStore.skills.find(s => s.id === skillId)
  const data = editingData.value[skillId]
  if (!skill || !data) return false

  // 比较分类
  if ((skill.category || '') !== data.category) return true

  // 比较备注
  if ((skill.notes || '') !== data.notes) return true

  // 比较标签数组
  const originalTags = skill.tags || []
  const editedTags = data.tags || []
  if (originalTags.length !== editedTags.length) return true
  for (let i = 0; i < originalTags.length; i++) {
    if (originalTags[i] !== editedTags[i]) return true
  }

  return false
}

onMounted(async () => {
  await skillStore.loadSkills()
  await skillStore.loadTools()
  await skillStore.loadCustomTools()
})

function showInstall() {
  router.push('/install')
}

function showSettings() {
  router.push('/settings')
}

function toggleExpand(skillId: string) {
  if (expandedSkill.value === skillId) {
    expandedSkill.value = null
  } else {
    expandedSkill.value = skillId
    // 初始化编辑状态
    const skill = skillStore.skills.find(s => s.id === skillId)
    if (skill) {
      editingData.value[skillId] = {
        category: skill.category || '',
        tags: [...(skill.tags || [])],
        notes: skill.notes || '',
        newTag: ''
      }
    }
  }
}

async function handleToggleTool(skillId: string, toolId: string) {
  try {
    await skillStore.toggleTool(skillId, toolId)
  } catch (e) {
    console.error('Toggle failed:', e)
  }
}

async function handleDeleteSkill(skillId: string) {
  const confirmed = showConfirm ? await showConfirm({
    title: '删除 Skill',
    message: '确定要删除这个 Skill 吗？\n此操作无法撤销。',
    type: 'danger',
    confirmText: '删除',
    cancelText: '取消'
  }) : confirm('确定要删除这个 Skill 吗？')

  if (!confirmed) {
    return
  }

  try {
    await skillStore.deleteSkill(skillId)
  } catch (e) {
    console.error('Delete failed:', e)
  }
}

async function handleUpdateSkill(skillId: string) {
  updatingSkill.value = skillId
  try {
    await skillStore.updateSkill(skillId)
    if (showNotification) {
      showNotification?.(`${skillId} 更新成功`, 'success')
    }
  } catch (e) {
    if (showNotification) {
      showNotification?.(`更新失败: ${e}`, 'error')
    }
  } finally {
    updatingSkill.value = null
  }
}

// 批量更新处理
async function handleBatchUpdate() {
  const confirmed = showConfirm ? await showConfirm({
    title: '批量更新',
    message: `确定要更新所有 ${skillStore.updatableSkillsCount} 个可更新的 Skill 吗？`,
    type: 'info',
    confirmText: '开始更新',
    cancelText: '取消'
  }) : confirm(`确定要更新所有 ${skillStore.updatableSkillsCount} 个可更新的 Skill 吗？`)

  if (!confirmed) {
    return
  }

  try {
    const result = await skillStore.batchUpdateAll()
    if (showNotification) {
      showNotification?.(`批量更新完成：成功 ${result.success} 个，失败 ${result.failed} 个`, result.failed > 0 ? 'warning' : 'success')
    }
  } catch (e) {
    if (showNotification) {
      showNotification?.(`批量更新失败: ${e}`, 'error')
    }
  }
}

// 切换忽略更新
async function handleToggleIgnoreUpdate(skillId: string) {
  skillStore.toggleIgnoreUpdate(skillId)
  const isIgnored = skillStore.ignoredUpdates.includes(skillId)
  if (showNotification) {
    showNotification?.(isIgnored ? '已忽略该 Skill 的更新' : '已恢复该 Skill 的更新提醒', 'info')
  }
}

// 获取分类图标类名
function getCategoryIconClass(category: string): string {
  const classes: Record<string, string> = {
    '内容创作': 'fa-pen-fancy text-pink-400',
    '认知增强': 'fa-lightbulb text-purple-400',
    '开发辅助': 'fa-code text-orange-400',
    '数据分析': 'fa-chart-bar text-cyan-400',
    '教育学习': 'fa-graduation-cap text-green-400',
    'AI/LLM': 'fa-brain text-blue-400',
    '其他': 'fa-cube text-gray-400'
  }
  return classes[category] || 'fa-cube text-gray-400'
}

// 获取分类背景类名
function getCategoryBgClass(category: string): string {
  const classes: Record<string, string> = {
    '内容创作': 'cat-content',
    '认知增强': 'cat-cognitive',
    '开发辅助': 'cat-dev',
    '数据分析': 'cat-data',
    '教育学习': 'cat-education',
    'AI/LLM': 'cat-ai',
    '其他': 'cat-other'
  }
  return classes[category] || 'cat-other'
}

// 获取侧边栏分类图标
function getSidebarCategoryIcon(category: string): string {
  const icons: Record<string, string> = {
    '全部': 'fa-folder',
    '内容创作': 'fa-pen-fancy',
    '认知增强': 'fa-lightbulb',
    '开发辅助': 'fa-code',
    '数据分析': 'fa-chart-bar',
    '教育学习': 'fa-graduation-cap',
    'AI/LLM': 'fa-brain'
  }
  return icons[category] || 'fa-cube'
}

// 获取标签颜色
function getTagColor(tag: string): string {
  const colors = [
    'bg-blue-500',
    'bg-purple-500',
    'bg-cyan-500',
    'bg-orange-500',
    'bg-pink-500',
    'bg-green-500'
  ]
  let hash = 0
  for (let i = 0; i < tag.length; i++) {
    hash = tag.charCodeAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
}

// 添加标签
function addTag(skillId: string) {
  const data = editingData.value[skillId]
  if (!data) return

  const tag = data.newTag.trim()
  if (tag && !data.tags.includes(tag)) {
    data.tags.push(tag)
  }
  data.newTag = ''
}

// 删除标签
function removeTag(skillId: string, tagIndex: number) {
  const data = editingData.value[skillId]
  if (!data) return

  data.tags.splice(tagIndex, 1)
}

// 保存元数据
async function saveMetadata(skillId: string) {
  const data = editingData.value[skillId]
  if (!data) {
    console.error('No editing data for skill:', skillId)
    return
  }

  console.log('Saving metadata:', skillId, data.category, data.tags, data.notes)

  try {
    await App.UpdateSkillMetadata(skillId, data.category, data.tags, data.notes)
    console.log('Metadata saved successfully')
    // 更新本地状态
    const skill = skillStore.skills.find(s => s.id === skillId)
    if (skill) {
      skill.category = data.category
      skill.tags = data.tags
      skill.notes = data.notes
    }
  } catch (e) {
    console.error('Save metadata failed:', e)
  }
}

// 只显示已启用的工具（用于 Stats Bar 和筛选）- 包含预置和自定义工具
const enabledTools = computed(() => skillStore.allEnabledTools)

// 只显示用户选择管理的工具（用于工具筛选和同步面板）- 包含预置和自定义工具
const managedTools = computed(() => skillStore.allEnabledTools)

const categories = computed(() => skillStore.categories)

// 计算常用标签（按使用频率排序，取前15个）
const popularTags = computed(() => {
  const tagCount: Record<string, number> = {}
  skillStore.skills.forEach(s => {
    s.tags?.forEach(t => {
      tagCount[t] = (tagCount[t] || 0) + 1
    })
  })
  return Object.entries(tagCount)
    .sort((a, b) => b[1] - a[1])
    .map(([tag]) => tag)
    .slice(0, 15)
})
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-cyber-dark via-cyber-panel to-cyber-dark text-gray-300 font-sans flex flex-col">
    <!-- Background Effects -->
    <div class="fixed inset-0 pointer-events-none overflow-hidden">
      <div class="absolute top-0 left-1/4 w-96 h-96 bg-cyber-accent/5 rounded-full blur-3xl animate-pulse-slow"></div>
      <div class="absolute bottom-0 right-1/4 w-96 h-96 bg-cyber-accent2/5 rounded-full blur-3xl animate-pulse-slow" style="animation-delay: 2s;"></div>
    </div>

    <!-- Header - 全宽 -->
    <header class="glass-panel sticky top-0 z-50 rounded-none">
      <div class="flex items-center justify-between px-6 py-4">
        <!-- Logo -->
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-[10px] bg-gradient-to-br from-cyber-accent to-cyber-accent2 flex items-center justify-center animate-glow">
            <i class="fas fa-cube text-white text-lg"></i>
          </div>
          <div>
            <h1 class="text-xl font-bold font-mono">
              <span class="gradient-text">Skill</span><span class="text-white">Hub</span>
            </h1>
            <p class="text-xs text-gray-500">AI Skill 管理器</p>
          </div>
        </div>

        <!-- Search Bar -->
        <div class="flex-1 max-w-xl mx-8">
          <div class="relative">
            <i class="fas fa-search absolute left-4 top-1/2 -translate-y-1/2 text-gray-500"></i>
            <input
              v-model="skillStore.searchQuery"
              type="text"
              placeholder="搜索 Skills..."
              class="w-full bg-cyber-dark border border-cyber-border rounded-xl py-2.5 pl-11 pr-4 text-sm focus:outline-none focus:border-cyber-accent transition-all"
            />
          </div>
        </div>

        <!-- Actions -->
        <div class="flex items-center gap-3">
          <button @click="showInstall" class="btn-secondary text-cyber-accent font-medium flex items-center gap-2">
            <i class="fas fa-plus"></i>
            <span>安装</span>
          </button>
          <button @click="showSettings" class="btn-icon">
            <i class="fas fa-cog"></i>
          </button>
        </div>
      </div>
    </header>

    <!-- Stats Bar - 全宽 -->
    <div class="border-b border-cyber-border bg-cyber-panel/30">
      <div class="px-6 py-3">
        <div class="flex items-center gap-8 text-sm">
          <div class="flex items-center gap-2">
            <span class="text-gray-500">总 Skills:</span>
            <span class="font-mono font-bold text-white text-lg">{{ skillStore.totalSkills }}</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-gray-500">已启用:</span>
            <span class="font-mono font-bold text-cyber-accent text-lg">{{ skillStore.enabledSkills }}</span>
          </div>
          <div class="h-4 w-px bg-cyber-border"></div>
          <div class="flex items-center gap-4">
            <span
              v-for="tool in enabledTools"
              :key="tool.id"
              class="flex items-center gap-1.5 text-xs text-gray-400"
            >
              <span class="status-dot active"></span>
              {{ tool.name }}
            </span>
          </div>
          <div class="flex-1"></div>
          <div class="flex items-center gap-4">
            <div
              class="flex items-center gap-1 cursor-pointer transition-all"
              @click="skillStore.filterUpdateOnly = !skillStore.filterUpdateOnly"
              :class="skillStore.filterUpdateOnly ? 'text-orange-400' : ''"
            >
              <span :class="skillStore.filterUpdateOnly ? 'text-orange-300' : 'text-gray-400'">可更新数：</span>
              <span class="font-semibold font-mono text-lg">{{ skillStore.updatableSkillsCount }}</span>
            </div>
            <button
              v-if="skillStore.updatableSkillsCount > 0 && !skillStore.batchUpdating"
              @click="handleBatchUpdate"
              class="btn-secondary text-xs px-3 py-1.5 text-cyber-accent border-cyber-accent/30"
            >
              <i class="fas fa-sync-alt mr-1"></i>一键更新
            </button>
            <div
              v-if="skillStore.batchUpdating"
              class="flex items-center gap-2 text-xs text-gray-400"
            >
              <i class="fas fa-spinner fa-spin"></i>
              <span>批量更新中...</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Content Area -->
    <div class="flex flex-1 overflow-hidden">
      <!-- Sidebar -->
      <aside class="w-[232px] border-r border-cyber-border bg-cyber-panel/50 flex-shrink-0 overflow-y-auto">
        <nav class="p-4">
          <!-- Categories -->
          <div class="mb-6">
            <p class="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3 px-3">分类</p>
            <ul class="space-y-1">
              <li v-for="cat in skillStore.categories" :key="cat">
                <a
                  href="#"
                  @click.prevent="skillStore.setCategory(cat)"
                  :class="[
                    'nav-item flex items-center gap-3 px-3 py-2 rounded-lg text-sm',
                    skillStore.selectedCategory === cat
                      ? 'active text-cyber-accent'
                      : 'text-gray-400 hover:text-gray-200'
                  ]"
                >
                  <i :class="['fas w-4', getSidebarCategoryIcon(cat)]"></i>
                  <span class="flex-1">{{ cat === '全部' ? '全部 Skills' : cat }}</span>
                  <span class="text-xs text-gray-500">
                    {{ cat === '全部' ? skillStore.totalSkills : skillStore.skills.filter(s => s.category === cat).length }}
                  </span>
                </a>
              </li>
            </ul>
          </div>

          <!-- Tool Filter -->
          <div class="mb-6">
            <p class="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3 px-3">工具筛选</p>
            <ul class="space-y-2">
              <li v-for="tool in managedTools" :key="tool.id">
                <label class="flex items-center gap-3 px-3 py-1.5 cursor-pointer hover:bg-cyber-dark/30 rounded-lg transition-all">
                  <input
                    type="checkbox"
                    class="cyber-checkbox"
                    :checked="skillStore.selectedTools.includes(tool.id)"
                    @change="skillStore.toggleToolFilter(tool.id)"
                  >
                  <span class="text-sm text-gray-400">{{ tool.name }}</span>
                </label>
              </li>
            </ul>
            <button
              v-if="skillStore.selectedTools.length > 0"
              @click="skillStore.clearToolFilter()"
              class="mt-2 text-xs text-gray-500 hover:text-cyber-accent px-3"
            >
              清除筛选
            </button>
          </div>

          <!-- Tags -->
          <div v-if="skillStore.allTags.length > 0">
            <p class="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3 px-3">标签</p>
            <div class="flex flex-wrap gap-1.5 px-2">
              <button
                v-for="tag in skillStore.allTags"
                :key="tag"
                @click="skillStore.toggleTag(tag)"
                :class="[
                  'flex items-center gap-1 px-1.5 py-1 rounded-[10px] text-xs transition-all',
                  skillStore.selectedTags.includes(tag)
                    ? 'bg-cyber-accent/15 border border-cyber-accent/40 text-cyber-accent'
                    : 'bg-cyber-panel/50 border border-cyber-border text-gray-400 hover:border-cyber-accent/30 hover:text-gray-300'
                ]"
              >
                <span :class="['w-1.5 h-1.5 rounded-full flex-shrink-0', getTagColor(tag)]"></span>
                <span>{{ tag }}</span>
              </button>
            </div>
            <button
              v-if="skillStore.selectedTags.length > 0"
              @click="skillStore.clearTagFilter()"
              class="mt-2 text-xs text-gray-500 hover:text-cyber-accent px-3"
            >
              清除筛选
            </button>
          </div>
        </nav>
      </aside>

      <!-- Main -->
      <main class="flex-1 overflow-y-auto p-6">
        <div class="max-w-4xl space-y-4">
          <!-- Loading -->
          <div v-if="skillStore.isLoading" class="text-center py-12">
            <div class="w-12 h-12 mx-auto border-4 border-cyber-border border-t-cyber-accent rounded-full animate-spin"></div>
            <p class="mt-4 text-gray-500">加载中...</p>
          </div>

          <!-- Empty State -->
          <div v-else-if="skillStore.filteredSkills.length === 0" class="text-center py-12">
            <div class="w-20 h-20 mx-auto mb-6 rounded-2xl bg-cyber-panel border border-cyber-border flex items-center justify-center">
              <span v-if="skillStore.filterUpdateOnly" class="text-3xl">🎉</span>
              <span v-else class="text-3xl">📦</span>
            </div>
            <h3 v-if="skillStore.filterUpdateOnly" class="text-lg font-semibold text-white mb-2">太棒了！</h3>
            <h3 v-else class="text-lg font-semibold text-white mb-2">暂无 Skills</h3>

            <p v-if="skillStore.filterUpdateOnly" class="text-sm text-gray-500 mb-4">所有 Skill 都是最新版本</p>
            <p v-else class="text-sm text-gray-500 mb-4">开始安装你的第一个 Skill 吧</p>

            <button
              v-if="skillStore.filterUpdateOnly"
              @click="skillStore.filterUpdateOnly = false"
              class="px-6 py-2.5 rounded-xl text-cyber-dark bg-gradient-to-r from-cyber-accent to-cyber-accent2 font-medium hover:opacity-90 transition-all"
            >
              查看全部 Skill
            </button>
            <button
              v-else
              @click="showInstall"
              class="px-6 py-2.5 rounded-xl text-cyber-dark bg-gradient-to-r from-cyber-accent to-cyber-accent2 font-medium hover:opacity-90 transition-all"
            >
              安装 Skill
            </button>
          </div>

          <!-- Skill Cards -->
          <div v-else class="space-y-4">
            <div
              v-for="skill in skillStore.filteredSkills"
              :key="skill.id"
              :class="[
                'skill-card glass-panel rounded-2xl p-5 cursor-pointer relative',
                skill.has_update && !skillStore.ignoredUpdates.includes(skill.id) ? 'has-update' : ''
              ]"
              @click="toggleExpand(skill.id)"
            >
              <!-- 更新徽章 -->
              <div
                v-if="skill.has_update && !skillStore.ignoredUpdates.includes(skill.id)"
                class="absolute -top-2 -right-2 px-3 py-1 rounded-full update-badge text-xs font-bold text-white z-20"
              >
                <i class="fas fa-arrow-up mr-1"></i>有更新
              </div>
              <div class="flex items-start justify-between relative z-10">
                <div class="flex-1">
                  <div class="flex items-center gap-3 mb-3">
                    <!-- Category Icon -->
                    <div :class="['w-12 h-12 rounded-[10px] flex items-center justify-center', getCategoryBgClass(skill.category)]">
                      <i :class="['fas text-xl', getCategoryIconClass(skill.category)]"></i>
                    </div>
                    <div>
                      <h3 class="text-lg font-semibold text-white font-mono">
                        {{ skill.original_name || skill.name }}
                      </h3>
                      <div class="flex items-center gap-3 mt-1">
                        <span class="text-xs text-gray-500" :title="skill.updated_at ? '更新时间' : '安装时间'">
                          <i class="far fa-calendar-alt mr-1"></i>{{ (skill.updated_at || skill.installed_at)?.substring(0, 10) || '-' }}
                        </span>
                        <span class="px-2 py-0.5 rounded-full bg-cyber-panel border border-cyber-border text-xs text-gray-400">
                          <i class="fas fa-folder mr-1"></i>{{ skill.category || '未分类' }}
                        </span>
                      </div>
                    </div>
                  </div>

                  <!-- Tags and Tool Status -->
                  <div class="flex items-center justify-between mb-3">
                    <div class="flex items-center gap-2">
                      <span
                        v-for="tag in skill.tags?.slice(0, 3)"
                        :key="tag"
                        class="tag px-2.5 py-1 rounded-lg text-xs text-cyber-accent"
                      >
                        {{ tag }}
                      </span>
                      <span v-if="skill.tags?.length > 3" class="text-xs text-gray-500">
                        +{{ skill.tags.length - 3 }}
                      </span>
                    </div>
                    <div class="flex items-center gap-3 text-xs">
                      <span
                        v-for="tool in managedTools"
                        :key="tool.id"
                        :class="[
                          'flex items-center gap-1.5',
                          skill.tools_enabled?.[tool.id] ? '' : 'text-gray-500'
                        ]"
                      >
                        <span
                          :class="[
                            'status-dot',
                            skill.tools_enabled?.[tool.id] ? 'active' : 'inactive'
                          ]"
                        ></span>
                        {{ tool.name }}
                      </span>
                    </div>
                  </div>
                </div>

                <!-- Actions -->
                <div class="flex items-center gap-2" @click.stop>
                  <button
                    @click="router.push(`/viewer/${skill.id}`)"
                    class="btn-icon"
                    title="查看"
                  >
                    <i class="fas fa-eye text-sm"></i>
                  </button>
                  <!-- 更新按钮（有更新时显示） -->
                  <button
                    v-if="skill.has_update && !skillStore.ignoredUpdates.includes(skill.id)"
                    @click="handleUpdateSkill(skill.id)"
                    :class="[
                      'btn-primary btn-primary-sm',
                      updatingSkill === skill.id ? 'opacity-50 cursor-not-allowed' : ''
                    ]"
                    title="更新可用"
                    :disabled="updatingSkill === skill.id"
                  >
                    <i :class="['fas', updatingSkill === skill.id ? 'fa-spinner fa-spin' : 'fa-arrow-up']"></i>
                  </button>
                  <!-- 忽略更新按钮 -->
                  <button
                    v-if="skill.has_update"
                    @click="handleToggleIgnoreUpdate(skill.id)"
                    class="btn-icon"
                    :title="skillStore.ignoredUpdates.includes(skill.id) ? '恢复更新提醒' : '忽略更新'"
                  >
                    <i :class="['fas text-sm', skillStore.ignoredUpdates.includes(skill.id) ? 'fa-bell-slash text-gray-500' : 'fa-bell text-orange-400']"></i>
                  </button>
                  <button
                    @click="handleDeleteSkill(skill.id)"
                    class="btn-icon delete"
                    title="删除"
                  >
                    <i class="fas fa-trash-alt text-sm"></i>
                  </button>
                  <button
                    class="btn-icon"
                    title="展开"
                    @click.stop="toggleExpand(skill.id)"
                  >
                    <i
                      :class="[
                        'fas fa-chevron-down text-sm transition-transform',
                        expandedSkill === skill.id ? 'rotate-180' : ''
                      ]"
                    ></i>
                  </button>
                </div>
              </div>

              <!-- Expand Panel -->
              <div
                v-if="expandedSkill === skill.id"
                class="pt-5 mt-5 border-t border-cyber-border"
                @click.stop
              >
                <!-- Skill 描述 -->
                <div class="mb-5">
                  <h4 class="text-sm font-semibold text-white mb-2 font-mono">
                    <i class="fas fa-align-left mr-2 text-cyber-accent"></i>Skill 描述
                  </h4>
                  <p class="text-sm text-gray-400 leading-relaxed">
                    {{ skill.description || '暂无描述' }}
                  </p>
                </div>

                <!-- 编辑元数据 & 同步到工具 -->
                <div class="grid grid-cols-2 gap-6">
                  <!-- Metadata Edit -->
                  <div>
                    <h4 class="text-sm font-semibold text-white mb-3 font-mono">
                      <i class="fas fa-edit mr-2 text-cyber-accent"></i>编辑元数据
                    </h4>
                    <div class="space-y-3">
                      <div>
                        <label class="text-xs text-gray-500 mb-1 block">分类</label>
                        <select
                          v-model="editingData[skill.id].category"
                          class="w-full bg-cyber-dark border border-cyber-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-cyber-accent"
                        >
                          <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
                        </select>
                      </div>
                      <div>
                        <label class="text-xs text-gray-500 mb-1 block">标签</label>
                        <div class="flex items-center gap-2 flex-wrap mb-2">
                          <span
                            v-for="(tag, tagIndex) in editingData[skill.id]?.tags"
                            :key="tagIndex"
                            class="tag px-2 py-1 rounded-lg text-xs text-cyber-accent flex items-center gap-1"
                          >
                            {{ tag }}
                            <i
                              class="fas fa-times cursor-pointer hover:text-red-400"
                              @click="removeTag(skill.id, tagIndex)"
                            ></i>
                          </span>
                        </div>
                        <div class="flex gap-2 mb-2">
                          <input
                            v-model="editingData[skill.id].newTag"
                            @keydown.enter.prevent="addTag(skill.id)"
                            type="text"
                            placeholder="输入标签后按回车..."
                            class="flex-1 bg-cyber-dark border border-cyber-border rounded-lg px-3 py-1.5 text-sm focus:outline-none focus:border-cyber-accent"
                          />
                          <button
                            @click="addTag(skill.id)"
                            :disabled="!editingData[skill.id]?.newTag?.trim()"
                            :class="[
                              'px-3 py-1.5 rounded-lg text-xs transition-all',
                              editingData[skill.id]?.newTag?.trim()
                                ? 'bg-cyber-accent/10 border border-cyber-accent/30 text-cyber-accent hover:bg-cyber-accent/20 cursor-pointer'
                                : 'bg-cyber-panel/50 border border-cyber-border text-gray-500 cursor-not-allowed'
                            ]"
                          >
                            <i class="fas fa-plus mr-1"></i>添加
                          </button>
                        </div>
                        <!-- 常用标签选择 -->
                        <div v-if="popularTags.length > 0" class="mb-1">
                          <p class="text-xs text-gray-500 mb-1">常用标签：</p>
                          <div class="flex flex-wrap gap-1">
                            <button
                              v-for="tag in popularTags"
                              :key="tag"
                              @click="() => {
                                const data = editingData[skill.id]
                                if (data && !data.tags?.includes(tag)) {
                                  if (!data.tags) data.tags = []
                                  data.tags.push(tag)
                                }
                              }"
                              :disabled="editingData[skill.id]?.tags?.includes(tag)"
                              :class="[
                                'px-2 py-1 rounded-lg text-xs transition-all flex items-center gap-1',
                                editingData[skill.id]?.tags?.includes(tag)
                                  ? 'bg-cyber-accent/15 border border-cyber-accent/40 text-cyber-accent opacity-50 cursor-not-allowed'
                                  : 'bg-cyber-panel/50 border border-cyber-border text-gray-400 hover:border-cyber-accent/30 hover:text-gray-300 cursor-pointer'
                              ]"
                            >
                              <span :class="['w-1.5 h-1.5 rounded-full flex-shrink-0', getTagColor(tag)]"></span>
                              {{ tag }}
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- Tool Sync + Notes -->
                  <div class="space-y-6">
                    <!-- Tool Sync -->
                    <div>
                      <h4 class="text-sm font-semibold text-white mb-3 font-mono">
                        <i class="fas fa-sync mr-2 text-cyber-accent2"></i>同步到工具
                      </h4>
                      <div class="grid grid-cols-2 gap-2">
                        <div
                          v-for="tool in managedTools"
                          :key="tool.id"
                          class="flex items-center justify-between p-2 rounded-lg bg-cyber-dark border border-cyber-border"
                        >
                          <div class="flex items-center gap-2">
                            <span class="text-xs font-bold text-orange-400" v-if="tool.id === 'claude-code'">C</span>
                            <span class="text-xs font-bold text-blue-400" v-else-if="tool.id === 'opencode'">O</span>
                            <span class="text-xs font-bold text-purple-400" v-else-if="tool.id === 'cursor'">Cu</span>
                            <span class="text-xs font-bold text-green-400" v-else-if="tool.id === 'codebuddy'">CB</span>
                            <span class="text-xs font-bold text-pink-400" v-else>T</span>
                            <span class="text-xs">{{ tool.name }}</span>
                          </div>
                          <button
                            @click.stop="handleToggleTool(skill.id, tool.id)"
                            :class="[
                              'toggle-switch',
                              { 'active': skill.tools_enabled?.[tool.id] }
                            ]"
                            :disabled="!tool.is_installed"
                          ></button>
                        </div>
                      </div>
                    </div>

                    <!-- Notes -->
                    <div class="space-y-3">
                      <div>
                        <label class="text-xs text-gray-500 mb-1 block">备注</label>
                        <textarea
                          v-model="editingData[skill.id].notes"
                          class="w-full bg-cyber-dark border border-cyber-border rounded-lg px-3 py-2 text-xs focus:outline-none focus:border-cyber-accent resize-none min-h-[120px]"
                          placeholder="添加备注..."
                        ></textarea>
                      </div>
                      <button
                        @click="saveMetadata(skill.id)"
                        :disabled="!hasChanges(skill.id)"
                        :class="[
                          'w-full py-2 rounded-lg text-sm font-medium transition-all',
                          hasChanges(skill.id)
                            ? 'bg-cyber-accent/10 border border-cyber-accent/30 text-cyber-accent hover:bg-cyber-accent/20 cursor-pointer'
                            : 'bg-cyber-panel/50 border border-cyber-border text-gray-500 cursor-not-allowed'
                        ]"
                      >
                        <i class="fas fa-save mr-2"></i>保存更改
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>
