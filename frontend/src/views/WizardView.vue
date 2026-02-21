<script setup lang="ts">
import { ref, onMounted, computed, watch, inject, type Ref } from 'vue'
import { useRouter } from 'vue-router'
import * as App from '../../wailsjs/go/backend/App'
import { useSettingsStore } from '../stores/settings'

const router = useRouter()
const settingsStore = useSettingsStore()

// 注入全局通知方法
const showNotification = inject<(message: string, type?: 'success' | 'error' | 'warning' | 'info', duration?: number) => void>('showNotification')

const currentStep: Ref<1 | 2> = ref<1 | 2>(1)
const storagePath = ref('~/.skill-hub/')
const loading = ref(false)

// 手动添加工具模态框
const showAddToolModal = ref(false)
const newToolName = ref('')
const newToolPath = ref('')
const addingTool = ref(false)

// 路径验证结果类型
interface PathValidation {
  is_valid: boolean
  is_writable: boolean
  disk_free_gb: number
  message: string
  status: string
}

const pathValidation = ref<PathValidation>({
  is_valid: false,
  is_writable: false,
  disk_free_gb: 0,
  message: '检查中...',
  status: 'valid'
})
const validating = ref(false)

// 工具信息类型
interface WizardTool {
  id: string
  name: string
  skills_path: string
  is_installed: boolean
  is_selected: boolean
  // UI 属性
  icon: string
  color: string
  badge: string
}

const tools = ref<WizardTool[]>([])

// 工具配置
const toolConfig: Record<string, { icon: string; color: string; badge: string }> = {
  'claude-code': { icon: 'C', color: 'orange', badge: '~/.claude/' },
  'opencode': { icon: 'O', color: 'blue', badge: '~/.config/opencode/' },
  'cursor': { icon: 'Cu', color: 'purple', badge: '~/.cursor/' },
  'codebuddy': { icon: 'CB', color: 'green', badge: '~/.codebuddy/' },
  'trae': { icon: 'T', color: 'pink', badge: '~/.trae/' }
}

// 加载检测到的工具
onMounted(async () => {
  try {
    // 尝试获取已保存的存储路径
    try {
      const savedPath = await App.GetSkillHubPath()
      if (savedPath && savedPath !== '') {
        storagePath.value = savedPath
      }
    } catch (e) {
      // 忽略错误，使用默认路径
      console.log('未获取到已保存路径，使用默认路径')
    }

    const detected = await App.GetDetectedTools()
    tools.value = detected.map(t => {
      const config = toolConfig[t.id] || { icon: '?', color: 'gray', badge: '未配置' }
      return {
        ...t,
        is_selected: t.is_installed, // 默认选中已检测的工具
        icon: config.icon,
        color: config.color,
        badge: config.badge
      }
    })
  } catch (e) {
    console.error('加载工具列表失败:', e)
    if (showNotification) {
      showNotification('加载工具列表失败: ' + e, 'error')
    }
  }

  // 初始验证路径
  await validatePath()
})

// 验证路径
let validateTimer: ReturnType<typeof setTimeout> | null = null
async function validatePath() {
  if (validateTimer) {
    clearTimeout(validateTimer)
  }

  // 防抖：500ms 后执行验证
  validateTimer = setTimeout(async () => {
    if (!storagePath.value || storagePath.value.trim() === '') {
      pathValidation.value = {
        is_valid: false,
        is_writable: false,
        disk_free_gb: 0,
        message: '请输入存储路径',
        status: 'error'
      }
      return
    }

    validating.value = true
    try {
      const result = await App.ValidateStoragePath(storagePath.value)
      pathValidation.value = result
    } catch (e) {
      pathValidation.value = {
        is_valid: false,
        is_writable: false,
        disk_free_gb: 0,
        message: '验证失败: ' + e,
        status: 'error'
      }
    } finally {
      validating.value = false
    }
  }, 500)
}

// 监听路径变化
watch(storagePath, () => {
  validatePath()
})

// 浏览文件夹（步骤1使用）
async function browseFolder() {
  try {
    const path = await App.SelectFolder()
    // 用户取消时 path 为空字符串，不做任何处理
    if (path && path !== '') {
      storagePath.value = path
    }
  } catch (e) {
    console.error('选择文件夹失败:', e)
  }
}

// 浏览文件夹（模态框使用）
async function browseFolderForTool() {
  try {
    const path = await App.SelectFolder()
    // 用户取消时 path 为空字符串，不做任何处理
    if (path && path !== '') {
      newToolPath.value = path
    }
  } catch (e) {
    console.error('选择文件夹失败:', e)
  }
}

// 切换工具选择
function toggleTool(toolId: string) {
  const tool = tools.value.find(t => t.id === toolId)
  if (tool) {
    tool.is_selected = !tool.is_selected
  }
}

// 打开手动添加工具模态框
function openAddToolModal() {
  newToolName.value = ''
  newToolPath.value = ''
  showAddToolModal.value = true
}

// 关闭模态框
function closeAddToolModal() {
  showAddToolModal.value = false
}

// 确认添加工具
async function confirmAddTool() {
  if (!newToolName.value.trim() || !newToolPath.value.trim()) {
    if (showNotification) {
      showNotification('请填写工具名称和路径', 'warning')
    }
    return
  }

  addingTool.value = true
  try {
    // 生成唯一 ID
    const customId = 'custom-' + Date.now()
    const newTool: WizardTool = {
      id: customId,
      name: newToolName.value.trim(),
      skills_path: newToolPath.value.trim(),
      is_installed: true,  // 手动添加的视为已安装
      is_selected: true,      // 默认选中
      icon: newToolName.value.trim().charAt(0).toUpperCase(),
      color: 'gray',
      badge: newToolPath.value.trim()
    }

    tools.value.push(newTool)
    closeAddToolModal()
    if (showNotification) {
      showNotification('工具添加成功', 'success')
    }
  } catch (e) {
    if (showNotification) {
      showNotification('添加工具失败: ' + e, 'error')
    }
  } finally {
    addingTool.value = false
  }
}

// 手动添加工具（旧函数保留）
function addCustomTool() {
  openAddToolModal()
}

// 获取工具颜色类
function getColorClasses(color: string): { bg: string; text: string } {
  const colors: Record<string, { bg: string; text: string }> = {
    orange: { bg: 'from-orange-500/20 to-orange-600/20', text: 'text-orange-400' },
    blue: { bg: 'from-blue-500/20 to-blue-600/20', text: 'text-blue-400' },
    purple: { bg: 'from-purple-500/20 to-purple-600/20', text: 'text-purple-400' },
    green: { bg: 'from-green-500/20 to-green-600/20', text: 'text-green-400' },
    pink: { bg: 'from-pink-500/20 to-pink-600/20', text: 'text-pink-400' },
    gray: { bg: 'from-gray-500/20 to-gray-600/20', text: 'text-gray-400' }
  }
  return colors[color] || colors.gray
}

// 完成设置
async function completeSetup() {
  loading.value = true
  try {
    const selectedTools = tools.value.filter(t => t.is_selected)

    // 分离预置工具和自定义工具
    const builtInTools = selectedTools
      .filter(t => !t.id.startsWith('custom-'))
      .map(t => t.id)

    const customTools = selectedTools
      .filter(t => t.id.startsWith('custom-'))
      .map(t => ({
        id: t.id,
        name: t.name,
        skills_path: t.skills_path,
        enabled: true,
        date_added: new Date().toISOString().split('T')[0]
      }))

    await App.InitializeWizard(storagePath.value, builtInTools, customTools)

    // 更新前端状态，标记为非首次运行并同步路径
    settingsStore.isFirstRun = false
    settingsStore.skillhubPath = storagePath.value

    router.push('/')
    if (showNotification) {
      showNotification('初始化设置成功', 'success')
    }
  } catch (e) {
    if (showNotification) {
      showNotification('初始化失败: ' + e, 'error')
    }
  } finally {
    loading.value = false
  }
}

// 返回上一步
function goToStep1() {
  currentStep.value = 1
}

// 进入下一步
function goToStep2() {
  currentStep.value = 2
}
</script>

<template>
  <!-- Background Glow Effects -->
  <div class="bg-glow bg-glow-1"></div>
  <div class="bg-glow bg-glow-2"></div>

  <!-- Main Container -->
  <div class="relative z-10 min-h-screen flex items-center justify-center p-6">
    <div class="w-full max-w-2xl animate-slide-up">
      <!-- Logo Header -->
      <div class="text-center mb-10">
        <div class="inline-flex items-center gap-3 mb-4">
          <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-cyber-accent to-cyber-accent2 flex items-center justify-center animate-glow">
            <i class="fas fa-cube text-white text-2xl"></i>
          </div>
        </div>
        <h1 class="text-3xl font-bold font-mono mb-2">
          <span class="gradient-text">Skill</span><span class="text-white">Hub</span>
        </h1>
        <p class="text-gray-500">AI 编程工具 Skill 管理器</p>
      </div>

      <!-- Wizard Card -->
      <div class="glass-panel rounded-3xl p-8">
        <!-- Step Indicators -->
        <div class="flex items-center justify-center gap-4 mb-8">
          <div class="flex items-center gap-3">
            <div
              :class="[
                'step-indicator w-12 h-12 rounded-full flex items-center justify-center font-bold text-sm transition-all bg-cyber-panel',
                currentStep >= 1 ? 'active' : ''
              ]"
            >
              <i v-if="currentStep > 1" class="fas fa-check text-cyber-accent"></i>
              <span v-show="currentStep <= 1">1</span>
            </div>
            <span :class="['text-sm font-medium', currentStep >= 1 ? 'text-white' : 'text-gray-500']">配置路径</span>
          </div>
          <div class="w-16 h-px bg-cyber-border"></div>
          <div class="flex items-center gap-3">
            <div
              :class="[
                'step-indicator w-12 h-12 rounded-full flex items-center justify-center font-bold text-sm transition-all bg-cyber-panel',
                currentStep >= 2 ? 'active' : ''
              ]"
            >
              <i v-if="currentStep > 2" class="fas fa-check text-cyber-accent"></i>
              <span v-show="currentStep <= 2">2</span>
            </div>
            <span :class="['text-sm font-medium', currentStep >= 2 ? 'text-white' : 'text-gray-500']">选择工具</span>
          </div>
        </div>

        <!-- Step 1: Configure Path -->
        <div v-show="currentStep === 1" class="step-content">
          <div class="text-center mb-8">
            <h2 class="text-2xl font-semibold text-white mb-2">配置存储位置</h2>
            <p class="text-sm text-gray-500">选择 SkillHub 的数据存储目录</p>
          </div>

          <div class="space-y-6 mb-6">
            <!-- 路径输入区域 -->
            <div>
              <label class="block text-sm font-medium text-gray-400 mb-3">SkillHub 存储路径</label>
              <div class="flex gap-3">
                <div class="flex-1 relative">
                  <i class="fas fa-folder absolute left-4 top-1/2 -translate-y-1/2 text-gray-500"></i>
                  <input
                    v-model="storagePath"
                    type="text"
                    class="path-input w-full py-3.5 pl-11 pr-4 text-sm font-mono"
                  >
                </div>
                <button
                  @click="browseFolder"
                  class="btn-secondary flex items-center gap-2"
                >
                  <i class="fas fa-folder-open"></i>
                  <span class="text-sm">浏览</span>
                </button>
              </div>
              <p class="mt-2 text-xs text-gray-500">
                <i class="fas fa-info-circle mr-1"></i>
                此目录将用于存储所有 Skill 文件和配置
              </p>
            </div>

            <!-- 路径验证 -->
            <div class="flex items-center gap-3 text-sm">
              <i
                v-if="validating"
                class="fas fa-spinner fa-spin text-gray-400"
              ></i>
              <i
                v-else-if="pathValidation.status === 'valid'"
                class="fas fa-check-circle text-cyber-accent"
              ></i>
              <i
                v-else-if="pathValidation.status === 'warning'"
                class="fas fa-exclamation-triangle text-yellow-400"
              ></i>
              <i
                v-else
                class="fas fa-times-circle text-red-400"
              ></i>
              <span :class="{
                'text-gray-400': pathValidation.status === 'valid',
                'text-yellow-400': pathValidation.status === 'warning',
                'text-red-400': pathValidation.status === 'error'
              }">{{ pathValidation.message }}</span>
            </div>
          </div>

          <!-- 导航按钮 -->
          <div class="flex justify-end">
            <button
              @click="goToStep2"
              :disabled="!pathValidation.is_valid || validating"
              :class="[
                'btn-primary flex items-center gap-2',
                !pathValidation.is_valid || validating ? 'opacity-50 cursor-not-allowed' : ''
              ]"
            >
              <span>{{ validating ? '验证中...' : '下一步' }}</span>
              <i v-if="!validating" class="fas fa-arrow-right"></i>
            </button>
          </div>
        </div>

        <!-- Step 2: Select Tools -->
        <div v-show="currentStep === 2" class="step-content">
          <div class="text-center mb-8">
            <h2 class="text-2xl font-semibold text-white mb-2">选择 AI 工具</h2>
            <p class="text-sm text-gray-500">选择要管理的 AI 编程工具（已自动检测）</p>
          </div>

          <!-- Tool List -->
          <div class="space-y-3 mb-6">
            <div v-for="tool in tools" :key="tool.id"
              :class="[
                'tool-card p-4 rounded-lg bg-cyber-dark border cursor-pointer transition-all',
                tool.is_selected ? 'selected' : 'border-cyber-border',
                !tool.is_installed && 'opacity-50 disabled'
              ]"
              @click="toggleTool(tool.id)"
            >
              <div class="flex items-center justify-between relative z-10">
                <div class="flex items-center gap-4">
                  <div :class="['w-10 h-10 rounded-lg flex items-center justify-center', getColorClasses(tool.color).bg]">
                    <span class="text-sm font-bold" :class="getColorClasses(tool.color).text">{{ tool.icon }}</span>
                  </div>
                  <div>
                    <h4 class="font-medium text-white">{{ tool.name }}</h4>
                    <p class="text-xs text-gray-500">{{ tool.skills_path }}</p>
                  </div>
                </div>
                <div class="flex items-center gap-3">
                  <span :class="['tag', tool.is_installed ? '' : 'muted']">
                    {{ tool.is_installed ? '已检测' : '未检测' }}
                  </span>
                  <div class="check-icon w-6 h-6 rounded-full bg-cyber-accent flex items-center justify-center">
                    <i class="fas fa-check text-cyber-dark text-xs"></i>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Add Custom Tool Button -->
          <button
            @click="addCustomTool"
            class="btn-dashed w-full flex items-center justify-center gap-2"
          >
            <i class="fas fa-plus"></i>
            <span class="text-sm">手动添加工具路径</span>
          </button>

          <!-- Navigation -->
          <div class="flex justify-between mt-8">
            <button
              @click="goToStep1"
              :disabled="loading"
              class="btn-ghost flex items-center gap-2"
            >
              <i class="fas fa-arrow-left"></i>
              <span>上一步</span>
            </button>
            <button
              @click="completeSetup"
              :disabled="loading || tools.filter(t => t.is_selected).length === 0"
              :class="[
                'btn-primary flex items-center gap-2',
                loading || tools.filter(t => t.is_selected).length === 0
                  ? 'disabled opacity-50 cursor-not-allowed'
                  : ''
              ]"
            >
              <span v-if="loading"><i class="fas fa-spinner fa-spin mr-2"></i>初始化中...</span>
              <span v-else>完成设置</span>
              <i v-if="!loading" class="fas fa-check"></i>
            </button>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="text-center mt-8 text-sm text-gray-600">
        <p>SkillHub v1.0 · 开源软件</p>
      </div>
    </div>

    <!-- Add Tool Modal -->
    <div v-if="showAddToolModal" class="modal-overlay" @click.self="closeAddToolModal">
      <div class="modal-content glass-panel">
        <!-- Modal Header -->
        <div class="modal-header">
          <h3 class="text-lg font-semibold text-white">手动添加工具</h3>
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
                @click="browseFolderForTool"
                class="btn-secondary flex items-center gap-2"
              >
                <i class="fas fa-folder-open"></i>
                <span class="text-sm">浏览</span>
              </button>
            </div>
            <p class="mt-2 text-xs text-gray-500">
              <i class="fas fa-info-circle mr-1"></i>
              指定该工具的 Skills 存储目录
            </p>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="modal-footer flex justify-end gap-3">
          <button
            @click="closeAddToolModal"
            :disabled="addingTool"
            class="btn-ghost flex items-center gap-2"
          >
            <span>取消</span>
          </button>
          <button
            @click="confirmAddTool"
            :disabled="addingTool"
            :class="[
              'btn-primary flex items-center gap-2',
              addingTool ? 'opacity-50 cursor-not-allowed' : ''
            ]"
          >
            <span v-if="addingTool"><i class="fas fa-spinner fa-spin mr-2"></i>添加中...</span>
            <span v-else>确认添加</span>
            <i v-if="!addingTool" class="fas fa-plus"></i>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Glass Panel Effect */
.glass-panel {
  position: relative;
  background: rgba(26, 26, 36, 0.6);
  backdrop-filter: blur(20px);
  border-radius: 24px;
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
      rgba(168, 85, 247, 0.25) 100%
  );
}

/* Gradient Text */
.gradient-text {
  background: linear-gradient(135deg, #00d4aa 0%, #a855f7 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* Step Indicators */
.step-indicator {
  transition: all 0.3s ease;
}

.step-indicator.active {
  background: linear-gradient(135deg, #00d4aa, #a855f7);
  box-shadow: 0 0 20px rgba(0, 212, 170, 0.4);
  color: #0a0a0f;
}

/* Input Field */
.path-input {
  background: rgba(10, 10, 15, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  transition: all 0.2s ease;
}

.path-input:hover {
  border-color: rgba(255, 255, 255, 0.15);
}

.path-input:focus {
  outline: none;
  border-color: #00d4aa;
  box-shadow: 0 0 0 3px rgba(0, 212, 170, 0.3),
              0 0 20px rgba(0, 212, 170, 0.1);
}

/* Tool Card */
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

.tool-card.selected {
  background: rgba(0, 212, 170, 0.08);
  border-color: #00d4aa;
}

.tool-card.selected::before {
  opacity: 1;
  background: linear-gradient(135deg,
      rgba(0, 212, 170, 0.5) 0%,
      rgba(168, 85, 247, 0.4) 100%
  );
}

.tool-card.disabled {
  opacity: 0.5;
}

.tool-card.disabled:hover {
  transform: none;
}

/* Check Icon */
.check-icon {
  opacity: 0;
  transform: scale(0);
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.tool-card.selected .check-icon {
  opacity: 1;
  transform: scale(1);
}

/* Tag */
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

/* Buttons */
button {
  border: none;
  cursor: pointer;
  transition: all 0.3s ease;
  font-family: inherit;
}

button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Background Effects */
.bg-glow-1 {
  position: fixed;
  width: 600px;
  height: 600px;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.15;
  pointer-events: none;
  background: #00d4aa;
  animation: float 8s ease-in-out infinite;
}

.bg-glow-2 {
  position: fixed;
  width: 600px;
  height: 600px;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.15;
  pointer-events: none;
  background: #a855f7;
  animation: float 8s ease-in-out infinite reverse;
}

@keyframes float {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(30px, -30px); }
}

/* Animation */
.animate-slide-up {
  animation: slideUp 0.5s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.animate-glow {
  animation: glow 3s ease-in-out infinite alternate;
}

@keyframes glow {
  from {
    box-shadow: 0 0 20px rgba(0, 212, 170, 0.1);
  }
  to {
    box-shadow: 0 0 40px rgba(0, 212, 170, 0.3), 0 0 60px rgba(168, 85, 247, 0.1);
  }
}

/* Background gradient */
body {
  background: linear-gradient(135deg, #0a0a0f 0%, #12121a 50%, #0f0f16 100%);
}

/* Info Panel */
.info-panel {
  position: relative;
  background: rgba(10, 10, 15, 0.6);
  border-radius: 12px;
}

.info-panel::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  padding: 1px;
  background: linear-gradient(135deg,
      rgba(255, 255, 255, 0.06) 0%,
      rgba(255, 255, 255, 0.02) 50%,
      rgba(255, 255, 255, 0.06) 100%
  );
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  pointer-events: none;
}

/* Secondary Button - 匹配 HTML 原型 */
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
  box-shadow: 0 0 16px rgba(0, 212, 170, 0.1);
}

/* Primary Button - 匹配 HTML 原型 */
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

/* Ghost Button - 匹配 HTML 原型 */
.btn-ghost {
  background: rgba(18, 18, 26, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #9ca3af;
  padding: 10px 16px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}
.btn-ghost:hover {
  border-color: rgba(0, 212, 170, 0.4);
  color: #00d4aa;
  background: rgba(0, 212, 170, 0.08);
}

/* Dashed Button - 匹配 HTML 原型 */
.btn-dashed {
  background: transparent;
  border: 1px dashed rgba(255, 255, 255, 0.15);
  color: #9ca3af;
  padding: 16px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}
.btn-dashed:hover {
  border-color: rgba(0, 212, 170, 0.4);
  color: #00d4aa;
  background: rgba(0, 212, 170, 0.05);
}

/* ==================== Modal Styles ==================== */
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

/* Modal Content */
.modal-content {
  width: 100%;
  max-width: 480px;
  padding: 24px;
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

/* Modal Header */
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.modal-close {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #9ca3af;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
}

.modal-close:hover {
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.3);
  color: #ef4444;
}

/* Modal Body */
.modal-body {
  margin-bottom: 24px;
}

/* Modal Input */
.modal-input {
  background: rgba(10, 10, 15, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  color: #e5e7eb;
  transition: all 0.2s ease;
}

.modal-input:hover {
  border-color: rgba(255, 255, 255, 0.15);
}

.modal-input:focus {
  outline: none;
  border-color: #00d4aa;
  box-shadow: 0 0 0 3px rgba(0, 212, 170, 0.1);
}

.modal-input::placeholder {
  color: #6b7280;
}

/* Modal Footer */
.modal-footer {
  padding-top: 8px;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}
</style>
