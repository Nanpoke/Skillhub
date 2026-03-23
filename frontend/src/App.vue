<script setup lang="ts">
import { onMounted, onUnmounted, ref, provide } from 'vue'
import { useRouter } from 'vue-router'
import { useSettingsStore } from './stores/settings'
import { useSkillStore } from './stores/skills'
import { EventsOn } from '../wailsjs/runtime/runtime'
import ConfirmDialog from './components/ConfirmDialog.vue'

const router = useRouter()
const settingsStore = useSettingsStore()
const skillStore = useSkillStore()
const confirmDialogRef = ref<InstanceType<typeof ConfirmDialog>>()

// 全局通知系统
interface Notification {
  show: boolean
  message: string
  type: 'success' | 'error' | 'warning' | 'info'
  duration: number
}

const notification = ref<Notification>({
  show: false,
  message: '',
  type: 'info',
  duration: 3000
})

let notificationTimer: ReturnType<typeof setTimeout> | null = null

function showNotification(message: string, type: Notification['type'] = 'info', duration = 3000) {
  // 清除之前的定时器
  if (notificationTimer) {
    clearTimeout(notificationTimer)
  }

  // 显示通知
  notification.value = {
    show: true,
    message,
    type,
    duration
  }

  // 设置自动隐藏
  notificationTimer = setTimeout(() => {
    notification.value.show = false
  }, duration)
}

// 全局确认对话框
interface ConfirmOptions {
  title?: string
  message: string
  confirmText?: string
  cancelText?: string
  type?: 'info' | 'warning' | 'danger'
  details?: Record<string, string>
  extraOptions?: Array<{ label: string; value: string }>
}

async function showConfirm(options: string | ConfirmOptions): Promise<boolean | string> {
  if (!confirmDialogRef.value) {
    // 降级使用原生 confirm
    return confirm(typeof options === 'string' ? options : options.message)
  }

  const opts = typeof options === 'string'
    ? { title: '确认操作', message: options, type: 'warning' as const }
    : { title: '确认操作', type: 'warning' as const, ...options }

  return await confirmDialogRef.value.open(opts)
}

// 提供全局方法
provide('showNotification', showNotification)
provide('showConfirm', showConfirm)

// 获取通知图标
function getNotificationIcon(type: Notification['type']): string {
  const icons = {
    success: 'fas fa-check-circle',
    error: 'fas fa-times-circle',
    warning: 'fas fa-exclamation-triangle',
    info: 'fas fa-info-circle'
  }
  return icons[type]
}

onMounted(async () => {
  await settingsStore.loadSettings()

  // Check if first run - show wizard
  if (settingsStore.isFirstRun) {
    router.push('/wizard')
  }

  // 监听更新检查事件
  EventsOn('updates:started', () => {
    console.log('更新检查开始')
  })

  EventsOn('updates:completed', async (data: any) => {
    console.log('更新检查完成:', data)
    // 重新加载 Skills 数据以获取最新的 has_update 状态
    await skillStore.loadSkills()
    if (data.update_count > 0) {
      showNotification(`发现 ${data.update_count} 个 Skill 有更新`, 'info', 5000)
    }
  })

  EventsOn('updates:failed', (data: any) => {
    console.error('更新检查失败:', data.error)
    // 不显示错误通知，避免打扰用户
  })
})

onUnmounted(() => {
  // 清理事件监听器（如果需要）
})
</script>

<template>
  <div class="app-container">
    <router-view />

    <!-- 全局通知横幅 -->
    <Transition name="slide-down">
      <div
        v-if="notification.show"
        class="notification-banner"
        :class="`banner-${notification.type}`"
      >
        <i :class="getNotificationIcon(notification.type)"></i>
        <span>{{ notification.message }}</span>
      </div>
    </Transition>

    <!-- 全局确认对话框 -->
    <ConfirmDialog ref="confirmDialogRef" />
  </div>
</template>

<style scoped>
.app-container {
  position: relative;
}

/* 通知横幅样式 */
.notification-banner {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1000;
  padding: 12px 24px;
  border-radius: 8px;
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  border: 1px solid;
}

/* 不同类型通知的样式 */
.banner-success {
  background: linear-gradient(135deg, rgba(34, 197, 94, 0.15), rgba(22, 163, 74, 0.1));
  border-color: rgba(34, 197, 94, 0.3);
  color: #bbf7d0;
}

.banner-error {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.15), rgba(220, 38, 38, 0.1));
  border-color: rgba(239, 68, 68, 0.3);
  color: #fecaca;
}

.banner-warning {
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.15), rgba(245, 158, 11, 0.1));
  border-color: rgba(251, 191, 36, 0.3);
  color: #ffedd5;
}

.banner-info {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.15), rgba(37, 99, 235, 0.1));
  border-color: rgba(59, 130, 246, 0.3);
  color: #dbeafe;
}

/* 滑入滑出动画 */
.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.3s ease-out;
}

.slide-down-enter-from {
  opacity: 0;
  transform: translate(-50%, -100%);
}

.slide-down-enter-to {
  opacity: 1;
  transform: translate(-50%, 0);
}

.slide-down-leave-from {
  opacity: 1;
  transform: translate(-50%, 0);
}

.slide-down-leave-to {
  opacity: 0;
  transform: translate(-50%, -100%);
}
</style>

<style>
/* ==================== 更新徽章 ==================== */
.update-badge {
  background: linear-gradient(135deg, #00d4aa, #a855f7);
  animation: pulse 2s infinite;
}

/* ==================== Skill卡片定位 ==================== */
.skill-card {
  position: relative;
}

/* ==================== 脉冲动画 ==================== */
@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.8;
    transform: scale(1.05);
  }
}

/* ==================== 有更新的卡片效果 ==================== */
.skill-card.has-update {
  border-color: rgba(0, 212, 170, 0.3);
  box-shadow: 0 0 20px rgba(0, 212, 170, 0.1);
}

.skill-card.has-update:hover {
  border-color: rgba(0, 212, 170, 0.5);
  box-shadow: 0 0 30px rgba(0, 212, 170, 0.15);
}

/* ==================== 主按钮 ==================== */
.btn-primary {
  background: linear-gradient(135deg, #00d4aa, #a855f7);
  color: #0a0a0f;
  padding: 8px 16px;
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

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}

.btn-primary-sm {
  width: 36px !important;
  height: 36px !important;
  padding: 0 !important;
  font-size: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}
</style>

