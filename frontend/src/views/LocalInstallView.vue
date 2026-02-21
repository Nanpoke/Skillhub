<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import * as App from '../../wailsjs/go/backend/App'
import { skill } from '../../wailsjs/go/models'

const router = useRouter()

const isLoading = ref(false)
const error = ref<string | null>(null)
const success = ref(false)

// 安装选项
const installOptions = ref<skill.InstallOptions>({
  category: '',
  tags: [],
  notes: ''
})

// 选中的路径
const selectedPath = ref('')

const categories = ['内容创作', '认知增强', '开发辅助', '数据分析', '教育学习', 'AI/LLM', '其他']

function goBack() {
  router.push('/install')
}

async function selectFolder() {
  // TODO: 使用 Wails 的文件对话框
  // 目前使用简单的输入框
  const path = prompt('请输入 Skill 文件夹路径:')
  if (path) {
    selectedPath.value = path
    scanPath(path)
  }
}

async function scanPath(path: string) {
  isLoading.value = true
  error.value = null

  try {
    const info = await App.ScanSkillPath(path)
    if (info) {
      // 显示 Skill 信息
      console.log('Skill info:', info)
    }
  } catch (e) {
    error.value = String(e)
  } finally {
    isLoading.value = false
  }
}

async function install() {
  if (!selectedPath.value) {
    error.value = '请先选择 Skill 文件夹'
    return
  }

  isLoading.value = true
  error.value = null

  try {
    await App.InstallFromPath(selectedPath.value, installOptions.value)
    success.value = true
  } catch (e) {
    error.value = String(e)
  } finally {
    isLoading.value = false
  }
}

function goHome() {
  router.push('/')
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-cyber-dark via-cyber-panel to-cyber-dark text-gray-300 font-sans">
    <!-- Header -->
    <header class="glass-panel sticky top-0 z-50 border-b border-cyber-border">
      <div class="max-w-4xl mx-auto px-6 py-4">
        <div class="flex items-center gap-4">
          <button @click="goBack" class="p-2 rounded-lg bg-cyber-panel border border-cyber-border hover:border-cyber-accent/50 transition-all">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <div>
            <h1 class="text-xl font-bold font-mono">
              <span class="gradient-text">本地</span> <span class="text-white">安装</span>
            </h1>
            <p class="text-xs text-gray-500">从本地文件夹安装 Skill</p>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-4xl mx-auto p-6">
      <div v-if="!success" class="space-y-6">
        <!-- Drop Zone -->
        <div class="glass-panel rounded-2xl p-8 border border-cyber-border border-dashed text-center">
          <div class="w-20 h-20 mx-auto mb-4 rounded-2xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center">
            <span class="text-4xl">📁</span>
          </div>
          <h3 class="text-lg font-semibold text-white mb-2">选择 Skill 文件夹</h3>
          <p class="text-sm text-gray-500 mb-4">文件夹中必须包含 SKILL.md 文件</p>
          <button
            @click="selectFolder"
            class="px-6 py-3 rounded-xl bg-gradient-to-r from-cyber-accent to-cyber-accent2 text-cyber-dark font-semibold hover:opacity-90 transition-all"
          >
            选择文件夹
          </button>
        </div>

        <!-- Selected Path -->
        <div v-if="selectedPath" class="glass-panel rounded-2xl p-6 border border-cyber-border">
          <h4 class="text-sm font-semibold text-gray-400 mb-3">已选择</h4>
          <code class="block p-3 rounded-xl bg-cyber-dark text-sm font-mono text-cyber-accent break-all">
            {{ selectedPath }}
          </code>
        </div>

        <!-- Install Options -->
        <div v-if="selectedPath" class="glass-panel rounded-2xl p-6 border border-cyber-border">
          <h4 class="text-sm font-semibold text-gray-400 mb-4">安装选项</h4>
          <div class="space-y-4">
            <div>
              <label class="text-xs text-gray-500 block mb-2">分类</label>
              <select
                v-model="installOptions.category"
                class="w-full bg-cyber-dark border border-cyber-border rounded-xl py-3 px-4 text-sm focus:outline-none focus:border-cyber-accent"
              >
                <option value="">选择分类</option>
                <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
              </select>
            </div>
            <div>
              <label class="text-xs text-gray-500 block mb-2">备注</label>
              <textarea
                v-model="installOptions.notes"
                placeholder="添加备注..."
                rows="3"
                class="w-full bg-cyber-dark border border-cyber-border rounded-xl py-3 px-4 text-sm focus:outline-none focus:border-cyber-accent resize-none"
              ></textarea>
            </div>
          </div>
        </div>

        <!-- Error -->
        <div v-if="error" class="p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
          {{ error }}
        </div>

        <!-- Actions -->
        <div v-if="selectedPath" class="flex gap-4">
          <button
            @click="goBack"
            class="flex-1 py-3 rounded-xl bg-cyber-panel border border-cyber-border text-gray-300 font-medium hover:border-cyber-accent/50 transition-all"
          >
            取消
          </button>
          <button
            @click="install"
            :disabled="isLoading"
            class="flex-1 py-3 rounded-xl bg-gradient-to-r from-cyber-accent to-cyber-accent2 text-cyber-dark font-semibold hover:opacity-90 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
          >
            <div v-if="isLoading" class="w-5 h-5 border-2 border-cyber-dark/30 border-t-cyber-dark rounded-full animate-spin"></div>
            <span v-else>安装</span>
          </button>
        </div>
      </div>

      <!-- Success -->
      <div v-else class="text-center py-12">
        <div class="w-20 h-20 mx-auto mb-6 rounded-full bg-cyber-accent/20 flex items-center justify-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-cyber-accent" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <h3 class="text-xl font-semibold text-white mb-2">安装完成！</h3>
        <p class="text-sm text-gray-500 mb-6">Skill 已成功安装</p>
        <button
          @click="goHome"
          class="px-8 py-3 rounded-xl bg-gradient-to-r from-cyber-accent to-cyber-accent2 text-cyber-dark font-semibold hover:opacity-90 transition-all"
        >
          返回主页
        </button>
      </div>
    </main>
  </div>
</template>
