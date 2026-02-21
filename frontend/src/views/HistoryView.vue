<script setup lang="ts">
import { ref, onMounted, computed, inject } from 'vue'
import { useRouter } from 'vue-router'
import * as App from '../../wailsjs/go/backend/App'
import type { skill } from '../../wailsjs/go/models'

const router = useRouter()

// 操作日志类型
type OperationLog = skill.OperationLog
type SkillStatus = skill.SkillStatus

const logs = ref<OperationLog[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

// 筛选器
const filterAction = ref<string>('all')
const filterSource = ref<string>('all')

// 注入全局方法
const showNotification = inject<(message: string, type?: 'success' | 'error' | 'warning' | 'info', duration?: number) => void>('showNotification')
const showConfirm = inject<(options: any) => Promise<boolean>>('showConfirm')

// 加载操作日志
async function loadLogs() {
  loading.value = true
  error.value = null

  try {
    const result = await App.GetOperationLogs()
    logs.value = result
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

// 清空日志
async function clearLogs() {
  const confirmed = showConfirm ? await showConfirm({
    title: '清空操作日志',
    message: '确定要清空所有操作日志吗？\n此操作无法撤销。',
    type: 'danger',
    confirmText: '清空',
    cancelText: '取消'
  }) : confirm('确定要清空所有操作日志吗？')

  if (!confirmed) {
    return
  }

  try {
    await App.ClearOperationLogs()
    logs.value = []
    if (showNotification) {
      showNotification('历史日志已清空', 'success')
    }
  } catch (e) {
    if (showNotification) {
      showNotification('清空失败: ' + e, 'error')
    }
  }
}

// 返回
function goBack() {
  router.back()
}

// 格式化时间
function formatTime(timestamp: any): string {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) {
    return `${days} 天前`
  } else if (hours > 0) {
    return `${hours} 小时前`
  } else if (minutes > 0) {
    return `${minutes} 分钟前`
  } else {
    return '刚刚'
  }
}

// 格式化完整时间
function formatFullTime(timestamp: any): string {
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 获取操作图标
function getActionIcon(action: string): string {
  const icons: Record<string, string> = {
    'install': 'fa-download text-green-400',
    'uninstall': 'fa-trash text-red-400',
    'delete': 'fa-trash text-red-400',
    'update': 'fa-sync text-blue-400',
    'enable': 'fa-toggle-on text-cyber-accent',
    'disable': 'fa-toggle-off text-gray-400',
    'metadata': 'fa-edit text-purple-400'
  }
  return icons[action] || 'fa-terminal text-gray-400'
}

// 获取操作文本
function getActionText(action: string): string {
  const texts: Record<string, string> = {
    'install': '安装',
    'uninstall': '卸载',
    'delete': '删除',
    'update': '更新',
    'enable': '启用',
    'disable': '禁用',
    'metadata': '修改元数据'
  }
  return texts[action] || action
}

// 获取来源文本
function getSourceText(source: string): string {
  const texts: Record<string, string> = {
    'local': '本地',
    'git': 'Git 仓库',
    'skills.sh': 'skills.sh'
  }
  return texts[source] || source
}

// 获取状态颜色
function getStatusColor(status: string): string {
  const colors: Record<string, string> = {
    'success': 'text-green-400',
    'failed': 'text-red-400',
    'pending': 'text-yellow-400'
  }
  return colors[status] || 'text-gray-400'
}

// 获取状态文本
function getStatusText(status: string): string {
  const texts: Record<string, string> = {
    'success': '成功',
    'failed': '失败',
    'pending': '进行中'
  }
  return texts[status] || status
}

// 筛选后的日志
const filteredLogs = computed(() => {
  return logs.value.filter(log => {
    if (filterAction.value !== 'all' && log.action !== filterAction.value) {
      return false
    }
    if (filterSource.value !== 'all' && log.source !== filterSource.value) {
      return false
    }
    return true
  })
})

// 可用的操作类型
const availableActions = computed(() => {
  const actions = new Set(logs.value.map(log => log.action))
  return Array.from(actions)
})

// 可用的来源
const availableSources = computed(() => {
  const sources = new Set(logs.value.map(log => log.source))
  return Array.from(sources)
})

// 统计信息
const stats = computed(() => {
  const total = logs.value.length
  const installCount = logs.value.filter(l => l.action === 'install').length
  const uninstallCount = logs.value.filter(l => l.action === 'uninstall').length
  const updateCount = logs.value.filter(l => l.action === 'update').length

  return { total, installCount, uninstallCount, updateCount }
})

onMounted(() => {
  loadLogs()
})
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-cyber-dark via-cyber-panel to-cyber-dark text-gray-300 font-sans">
    <!-- Header -->
    <header class="glass-panel sticky top-0 z-50 border-b border-cyber-border">
      <div class="max-w-5xl mx-auto px-6 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <button @click="goBack" class="p-2 rounded-lg bg-cyber-panel border border-cyber-border hover:border-cyber-accent/50 transition-all">
              <i class="fas fa-arrow-left text-gray-400"></i>
            </button>
            <div>
              <h1 class="text-xl font-bold font-mono">
                <span class="gradient-text">操作</span> 历史
              </h1>
              <p class="text-xs text-gray-500 mt-1">最近 10 天的操作记录</p>
            </div>
          </div>
          <button
            @click="clearLogs"
            class="px-4 py-2 rounded-lg text-sm border border-red-500/30 text-red-400 hover:bg-red-500/10 transition-all"
          >
            <i class="fas fa-trash-alt mr-2"></i>清空
          </button>
        </div>
      </div>
    </header>

    <!-- Content -->
    <main class="max-w-5xl mx-auto p-6">
      <!-- Loading -->
      <div v-if="loading" class="text-center py-12">
        <div class="w-12 h-12 mx-auto border-4 border-cyber-border border-t-cyber-accent rounded-full animate-spin"></div>
        <p class="mt-4 text-gray-500">加载中...</p>
      </div>

      <!-- Error -->
      <div v-else-if="error" class="p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
        {{ error }}
      </div>

      <!-- Main Content -->
      <div v-else>
        <!-- Stats Cards -->
        <div class="grid grid-cols-4 gap-4 mb-6">
          <div class="glass-panel rounded-xl p-4 border border-cyber-border">
            <div class="text-2xl font-bold font-mono text-white">{{ stats.total }}</div>
            <div class="text-xs text-gray-500 mt-1">总操作数</div>
          </div>
          <div class="glass-panel rounded-xl p-4 border border-cyber-border">
            <div class="text-2xl font-bold font-mono text-green-400">{{ stats.installCount }}</div>
            <div class="text-xs text-gray-500 mt-1">安装</div>
          </div>
          <div class="glass-panel rounded-xl p-4 border border-cyber-border">
            <div class="text-2xl font-bold font-mono text-red-400">{{ stats.uninstallCount }}</div>
            <div class="text-xs text-gray-500 mt-1">卸载</div>
          </div>
          <div class="glass-panel rounded-xl p-4 border border-cyber-border">
            <div class="text-2xl font-bold font-mono text-blue-400">{{ stats.updateCount }}</div>
            <div class="text-xs text-gray-500 mt-1">更新</div>
          </div>
        </div>

        <!-- Filters -->
        <div class="glass-panel rounded-xl p-4 border border-cyber-border mb-6">
          <div class="flex items-center gap-4">
            <div class="flex-1">
              <label class="text-xs text-gray-500 mb-1 block">操作类型</label>
              <select
                v-model="filterAction"
                class="w-full bg-cyber-dark border border-cyber-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-cyber-accent"
              >
                <option value="all">全部</option>
                <option v-for="action in availableActions" :key="action" :value="action">
                  {{ getActionText(action) }}
                </option>
              </select>
            </div>
            <div class="flex-1">
              <label class="text-xs text-gray-500 mb-1 block">来源</label>
              <select
                v-model="filterSource"
                class="w-full bg-cyber-dark border border-cyber-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:border-cyber-accent"
              >
                <option value="all">全部</option>
                <option v-for="source in availableSources" :key="source" :value="source">
                  {{ getSourceText(source) }}
                </option>
              </select>
            </div>
          </div>
        </div>

        <!-- Empty State -->
        <div v-if="filteredLogs.length === 0" class="glass-panel rounded-xl p-12 border border-cyber-border text-center">
          <i class="fas fa-history text-5xl text-gray-600 mb-4"></i>
          <h3 class="text-lg font-semibold text-gray-400 mb-2">暂无操作记录</h3>
          <p class="text-sm text-gray-500">操作历史将在这里显示</p>
        </div>

        <!-- Timeline -->
        <div v-else class="relative">
          <!-- Timeline Line -->
          <div class="absolute left-8 top-0 bottom-0 w-px bg-cyber-border"></div>

          <!-- Log Items -->
          <div class="space-y-4">
            <div
              v-for="(log, index) in filteredLogs"
              :key="index"
              class="relative flex gap-4"
            >
              <!-- Timeline Dot -->
              <div class="relative z-10 flex-shrink-0">
                <div :class="['w-16 h-16 rounded-full flex items-center justify-center bg-cyber-panel border-2 border-cyber-border']">
                  <i :class="['fas text-xl', getActionIcon(log.action)]"></i>
                </div>
              </div>

              <!-- Log Content -->
              <div class="flex-1 glass-panel rounded-xl p-4 border border-cyber-border hover:border-cyber-accent/30 transition-all">
                <!-- Header -->
                <div class="flex items-start justify-between mb-3">
                  <div>
                    <h3 class="text-sm font-semibold text-white">
                      <i :class="['fas mr-2', getActionIcon(log.action)]"></i>
                      {{ getActionText(log.action) }}
                    </h3>
                    <p class="text-xs text-gray-500 mt-1">
                      <span :class="['px-2 py-0.5 rounded', log.source === 'git' ? 'bg-blue-500/20 text-blue-400' : log.source === 'skills.sh' ? 'bg-purple-500/20 text-purple-400' : 'bg-gray-500/20 text-gray-400']">
                        {{ getSourceText(log.source) }}
                      </span>
                    </p>
                  </div>
                  <div class="text-right">
                    <div class="text-sm text-gray-400">{{ formatTime(log.timestamp) }}</div>
                    <div class="text-xs text-gray-600 mt-1" :title="formatFullTime(log.timestamp)">
                      {{ new Date(log.timestamp).toLocaleDateString() }}
                    </div>
                  </div>
                </div>

                <!-- Skills List -->
                <div v-if="log.skills && log.skills.length > 0" class="space-y-1">
                  <div
                    v-for="(skill, skillIndex) in log.skills"
                    :key="skillIndex"
                    class="flex items-center justify-between py-1.5 px-3 rounded-lg bg-cyber-dark"
                  >
                    <span class="text-sm font-mono text-gray-300">{{ skill.name }}</span>
                    <span :class="['text-xs font-medium', getStatusColor(skill.status)]">
                      {{ getStatusText(skill.status) }}
                    </span>
                  </div>
                </div>

                <!-- Duration -->
                <div v-if="log.duration_ms" class="mt-2 pt-2 border-t border-cyber-border">
                  <span class="text-xs text-gray-500">
                    <i class="fas fa-clock mr-1"></i>
                    耗时 {{ (log.duration_ms / 1000).toFixed(2) }} 秒
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>
