<script setup lang="ts">
import { ref, inject, computed } from 'vue'

// 注入全局通知方法
const showNotification = inject<(message: string, type?: 'success' | 'error' | 'warning' | 'info', duration?: number) => void>('showNotification')

interface ConfirmDialogOptions {
  title?: string
  message: string
  confirmText?: string
  cancelText?: string
  type?: 'info' | 'warning' | 'danger'
  details?: Record<string, string>
  extraOptions?: Array<{ label: string; value: string }>
}

const visible = ref(false)
const loading = ref(false)
const options = ref<Omit<Required<ConfirmDialogOptions>, 'details' | 'extraOptions'> & {
  details?: Record<string, string>
  extraOptions?: Array<{ label: string; value: string }>
}>({
  title: '确认操作',
  message: '您确定要执行此操作吗？',
  confirmText: '确定',
  cancelText: '取消',
  type: 'info'
})

let resolvePromise: ((value: boolean | string) => void) | null = null

function open(opts: ConfirmDialogOptions): Promise<boolean | string> {
  options.value = {
    title: '确认操作',
    confirmText: '确定',
    cancelText: '取消',
    type: 'info',
    details: undefined,
    extraOptions: undefined,
    ...opts
  }
  visible.value = true

  return new Promise((resolve) => {
    resolvePromise = resolve
  })
}

function close() {
  visible.value = false
  loading.value = false
}

function handleConfirm() {
  if (resolvePromise) {
    resolvePromise(true)
  }
  close()
}

function handleCancel() {
  if (resolvePromise) {
    resolvePromise(false)
  }
  close()
}

function handleExtraOption(value: string) {
  if (resolvePromise) {
    resolvePromise(value)
  }
  close()
}

// 计算是否有详细信息
const hasDetails = computed(() => {
  return options.value.details && Object.keys(options.value.details).length > 0
})

// 计算是否有额外选项
const hasExtraOptions = computed(() => {
  return options.value.extraOptions && options.value.extraOptions.length > 0
})

// 提供 open 方法给父组件
defineExpose({
  open
})
</script>

<template>
  <Transition name="fade">
    <div v-if="visible" class="modal-overlay" @click.self="handleCancel">
      <Transition name="modal-slide">
        <div class="confirm-dialog glass-panel" v-if="visible">
          <!-- Header -->
          <div class="dialog-header">
            <div class="flex items-center gap-3">
              <div class="dialog-icon" :class="`icon-${options.type}`">
                <i class="fas" :class="{
                  'fa-exclamation-circle': options.type === 'warning',
                  'fa-times-circle': options.type === 'danger',
                  'fa-question-circle': options.type === 'info'
                }"></i>
              </div>
              <h3 class="dialog-title">{{ options.title }}</h3>
            </div>
          </div>

          <!-- Body -->
          <div class="dialog-body">
            <p class="dialog-message" v-html="options.message"></p>

            <!-- Details Section -->
            <div v-if="hasDetails" class="dialog-details">
              <div v-for="(value, key) in options.details" :key="key" class="detail-row">
                <span class="detail-key">{{ key }}:</span>
                <span class="detail-value">{{ value }}</span>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="dialog-footer">
            <!-- Extra Options (if available) -->
            <template v-if="hasExtraOptions">
              <button
                v-for="opt in options.extraOptions"
                :key="opt.value"
                @click="handleExtraOption(opt.value)"
                class="btn-option"
                :class="`btn-option-${options.type}`"
                :disabled="loading"
              >
                {{ opt.label }}
              </button>
            </template>

            <!-- Standard Buttons (if no extra options) -->
            <template v-else>
              <button
                @click="handleCancel"
                class="btn-ghost"
                :disabled="loading"
              >
                {{ options.cancelText }}
              </button>
              <button
                @click="handleConfirm"
                class="btn-confirm"
                :class="`btn-${options.type}`"
                :disabled="loading"
              >
                <i v-if="loading" class="fas fa-spinner fa-spin mr-2"></i>
                {{ options.confirmText }}
              </button>
            </template>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>

<style scoped>
/* Modal Overlay */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* Confirm Dialog */
.confirm-dialog {
  width: 100%;
  max-width: 420px;
  padding: 0;
  border-radius: 16px;
  overflow: hidden;
  animation: modalSlideUp 0.3s ease-out;
}

@keyframes modalSlideUp {
  from {
    opacity: 0;
    transform: translateY(20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

/* Dialog Header */
.dialog-header {
  padding: 24px 24px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.dialog-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.dialog-icon.icon-info {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.2), rgba(37, 99, 235, 0.15));
  color: #60a5fa;
}

.dialog-icon.icon-warning {
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.2), rgba(245, 158, 11, 0.15));
  color: #fbbf24;
}

.dialog-icon.icon-danger {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.2), rgba(220, 38, 38, 0.15));
  color: #f87171;
}

.dialog-title {
  font-size: 18px;
  font-weight: 600;
  color: #ffffff;
  margin: 0;
}

/* Dialog Body */
.dialog-body {
  padding: 16px 24px 24px;
}

.dialog-message {
  font-size: 14px;
  line-height: 1.5;
  color: #d1d5db;
  margin: 0 0 12px 0;
  white-space: pre-line;
}

/* Dialog Details */
.dialog-details {
  margin-top: 12px;
  padding: 12px;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.detail-row {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
  font-size: 13px;
}

.detail-key {
  color: #9ca3af;
  font-weight: 500;
}

.detail-value {
  color: #f3f4f6;
  font-family: 'Consolas', 'Monaco', monospace;
}

/* Dialog Footer */
.dialog-footer {
  padding: 16px 24px 24px;
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  flex-wrap: wrap;
}

/* Buttons */
.btn-ghost {
  background: rgba(18, 18, 26, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #9ca3af;
  padding: 10px 20px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 14px;
  font-weight: 500;
}

.btn-ghost:hover {
  border-color: rgba(255, 255, 255, 0.2);
  color: #d1d5db;
  background: rgba(255, 255, 255, 0.05);
}

.btn-ghost:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-confirm {
  padding: 10px 20px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 14px;
  font-weight: 500;
  color: #0a0a0f;
  border: none;
}

.btn-confirm:hover:not(:disabled) {
  transform: translateY(-1px);
}

.btn-confirm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-info {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
}

.btn-info:hover {
  background: linear-gradient(135deg, #60a5fa, #3b82f6);
  box-shadow: 0 0 16px rgba(59, 130, 246, 0.3);
}

.btn-warning {
  background: linear-gradient(135deg, #f59e0b, #d97706);
}

.btn-warning:hover {
  background: linear-gradient(135deg, #fbbf24, #f59e0b);
  box-shadow: 0 0 16px rgba(245, 158, 11, 0.3);
}

.btn-danger {
  background: linear-gradient(135deg, #ef4444, #dc2626);
}

.btn-danger:hover {
  background: linear-gradient(135deg, #f87171, #ef4444);
  box-shadow: 0 0 16px rgba(239, 68, 68, 0.3);
}

/* Extra Option Buttons */
.btn-option {
  padding: 10px 16px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 14px;
  font-weight: 500;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(18, 18, 26, 0.8);
  color: #d1d5db;
}

.btn-option:hover {
  border-color: rgba(0, 212, 170, 0.4);
  color: #00d4aa;
  background: rgba(0, 212, 170, 0.08);
}

.btn-option-info:hover {
  border-color: rgba(59, 130, 246, 0.4);
  color: #60a5fa;
  background: rgba(59, 130, 246, 0.08);
}

.btn-option-warning:hover {
  border-color: rgba(245, 158, 11, 0.4);
  color: #fbbf24;
  background: rgba(245, 158, 11, 0.08);
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.modal-slide-enter-active {
  animation: modalSlideUp 0.3s ease-out;
}

.modal-slide-leave-active {
  animation: modalSlideUp 0.3s ease-out reverse;
}
</style>
