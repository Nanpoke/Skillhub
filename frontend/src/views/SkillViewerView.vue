<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import * as App from '../../wailsjs/go/backend/App'

const router = useRouter()

// 接收路由参数
const props = defineProps<{
  skillName?: string
}>()

// 数据状态
const files = ref<Array<{ name: string; path: string; is_dir: boolean; size: number; modified: string }>>([])
const selectedFile = ref<string | null>(null)
const fileContent = ref<string>('')
const loading = ref(false)
const error = ref<string | null>(null)

// 标签页状态
const activeTab = ref<'preview' | 'code'>('preview')

// 如果有 skillName，加载该技能的文件
onMounted(async () => {
  if (props.skillName) {
    await loadFiles(props.skillName)
  }
})

// 加载技能文件列表
async function loadFiles(skillName: string) {
  loading.value = true
  error.value = null

  try {
    const result = await App.ListSkillFiles(skillName)
    files.value = (result || []).map(f => ({
      name: String(f.name || ''),
      path: String(f.path || ''),
      is_dir: Boolean(f.is_dir || false),
      size: Number(f.size || 0),
      modified: String(f.modified || '')
    }))
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

// 选择文件
async function selectFile(filePath: string) {
  loading.value = true
  error.value = null

  try {
    const fileName = filePath.split('/').pop() || ''
    const content = await App.ReadSkillFile(props.skillName || '', fileName)
    fileContent.value = content
    selectedFile.value = filePath
  } catch (e) {
    error.value = String(e)
  } finally {
    loading.value = false
  }
}

// 返回
function goBack() {
  router.push('/')
}

// 获取文件图标
function getFileIcon(fileName: string, isDir: boolean): string {
  if (isDir) return 'fa-folder'

  const ext = fileName.split('.').pop()?.toLowerCase()
  switch (ext) {
    case 'md': return 'fa-file-markdown'
    case 'json': return 'fa-file-code'
    case 'txt': return 'fa-file-alt'
    case 'yml':
    case 'yaml': return 'fa-file-code'
    case 'ts':
    case 'js': return 'fa-file-code'
    case 'py': return 'fa-file-code'
    default: return 'fa-file'
  }
}

// 渲染 Markdown（简单实现）
function renderMarkdown(content: string): string {
  // 简单的 Markdown 转 HTML
  return content
    .replace(/^### (.*$)/gm, '<h3>$1</h3>')
    .replace(/^## (.*$)/gm, '<h2>$1</h2>')
    .replace(/^# (.*$)/gm, '<h1>$1</h1>')
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/`([^`]+)`/g, '<code class="bg-cyber-dark px-1 rounded text-cyber-accent">$1</code>')
    .replace(/\n/g, '<br>')
}

// 获取代码行数
const lineCount = computed(() => {
  return fileContent.value.split('\n').length
})

// 是否显示代码视图
const showCodeView = computed(() => {
  return selectedFile.value && activeTab.value === 'code'
})

// 渲染的 HTML 内容
const renderedContent = computed(() => {
  if (!fileContent.value) return ''

  if (activeTab.value === 'preview') {
    return renderMarkdown(fileContent.value)
  }

  // 代码视图
  return fileContent.value
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
                <span class="gradient-text">SKILL.md</span> 查看器
              </h1>
              <p class="text-xs text-gray-500 mt-1">{{ props.skillName }}</p>
            </div>
          </div>
          <button @click="goBack" class="p-2 rounded-lg hover:bg-cyber-panel transition-all">
            <i class="fas fa-times text-gray-400"></i>
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

      <!-- Viewer -->
      <div v-else class="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <!-- File Tree -->
        <div class="lg:col-span-1">
          <div class="glass-panel rounded-xl p-4 border border-cyber-border">
            <h3 class="text-sm font-semibold text-gray-400 mb-3">
              <i class="fas fa-folder-tree mr-2"></i>文件目录
            </h3>

            <!-- Empty State -->
            <div v-if="files.length === 0" class="text-center py-8">
              <i class="fas fa-folder-open text-4xl text-gray-600 mb-3"></i>
              <p class="text-sm text-gray-500">无文件</p>
            </div>

            <!-- File List -->
            <div v-else class="space-y-1">
              <div
                v-for="file in files"
                :key="file.path"
                @click="selectFile(file.path)"
                :class="[
                  'flex items-center gap-2 px-3 py-2 rounded-lg cursor-pointer transition-all text-sm',
                  selectedFile === file.path
                    ? 'bg-cyber-accent/10 border border-cyber-accent/30 text-cyber-accent'
                    : 'hover:bg-cyber-panel text-gray-400'
                ]"
              >
                <i :class="['fas', getFileIcon(file.name, file.is_dir)]" class="w-4"></i>
                <span class="truncate flex-1">{{ file.name }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- File Content -->
        <div class="lg:col-span-3">
          <!-- Empty State -->
          <div v-if="!selectedFile" class="glass-panel rounded-xl p-12 border border-cyber-border text-center">
            <i class="fas fa-file-code text-5xl text-gray-600 mb-4"></i>
            <h3 class="text-lg font-semibold text-gray-400 mb-2">未选择文件</h3>
            <p class="text-sm text-gray-500">请从左侧目录选择要查看的文件</p>
          </div>

          <!-- File View -->
          <div v-else class="glass-panel rounded-xl border border-cyber-border overflow-hidden">
            <!-- Toolbar -->
            <div class="flex items-center justify-between px-4 py-3 border-b border-cyber-border bg-cyber-panel/50">
              <div class="flex items-center gap-2">
                <i :class="['fas', getFileIcon(selectedFile, false)]" class="text-gray-500"></i>
                <span class="text-sm text-gray-400">{{ selectedFile?.split('/').pop() }}</span>
              </div>

              <!-- Tab Switch -->
              <div class="flex gap-1 bg-cyber-dark rounded-lg p-1">
                <button
                  @click="activeTab = 'preview'"
                  :class="[
                    'px-3 py-1 rounded text-xs font-medium transition-all',
                    activeTab === 'preview'
                      ? 'bg-cyber-accent text-white'
                      : 'text-gray-400 hover:text-white'
                  ]"
                >
                  <i class="fas fa-eye mr-1"></i>预览
                </button>
                <button
                  @click="activeTab = 'code'"
                  :class="[
                    'px-3 py-1 rounded text-xs font-medium transition-all',
                    activeTab === 'code'
                      ? 'bg-cyber-accent text-white'
                      : 'text-gray-400 hover:text-white'
                  ]"
                >
                  <i class="fas fa-code mr-1"></i>代码
                </button>
              </div>
            </div>

            <!-- Content Area -->
            <div class="p-4 min-h-[400px] max-h-[600px] overflow-auto">
              <!-- Code View -->
              <div v-if="activeTab === 'code'" class="font-mono text-sm">
                <pre class="text-gray-300"><code>{{ fileContent }}</code></pre>
              </div>

              <!-- Preview View -->
              <div v-else class="prose prose-invert max-w-none prose-sm">
                <div v-html="renderedContent"></div>
              </div>
            </div>

            <!-- Footer -->
            <div v-if="activeTab === 'code'" class="px-4 py-2 border-t border-cyber-border bg-cyber-panel/30 flex items-center justify-between text-xs text-gray-500">
              <span>{{ lineCount }} 行</span>
              <span>{{ files.find(f => f.path === selectedFile)?.size || 0 }} bytes</span>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>
