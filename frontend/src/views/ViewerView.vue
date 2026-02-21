<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import * as App from '../../wailsjs/go/backend/App'
import { marked } from 'marked'

const router = useRouter()
const route = useRoute()

const skillId = route.params.id as string
const currentTab = ref('files')
const loading = ref(false)
const error = ref<string | null>(null)

// 文件数据
const files = ref<Array<{ name: string; path: string; is_dir: boolean; size: number; modified: string }>>([])
const selectedFile = ref<string | null>(null)
const fileContent = ref<string>('')
const currentPath = ref('') // 当前目录的相对路径

// 配置 marked（使用 setOptions 静态方法）
marked.setOptions({
  breaks: true,
  gfm: true
})

// 加载文件列表
async function loadFiles() {
  loading.value = true
  error.value = null

  try {
    const result = await App.ListSkillFiles(skillId, currentPath.value)
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

// 组件加载时获取文件列表
onMounted(async () => {
  await loadFiles()
  // 默认选择 SKILL.md
  const skillMdFile = files.value.find(f => f.name.toLowerCase() === 'skill.md')
  if (skillMdFile) {
    await selectFile(skillMdFile.name, skillMdFile.is_dir)
  }
})

// 选择文件或进入目录
async function selectFile(fileName: string, isDir: boolean = false) {
  if (isDir) {
    // 进入目录
    if (currentPath.value) {
      currentPath.value = currentPath.value + '/' + fileName
    } else {
      currentPath.value = fileName
    }
    await loadFiles()
    selectedFile.value = null
    fileContent.value = ''
  } else {
    // 读取文件
    loading.value = true
    error.value = null

    try {
      // 构建相对路径
      const relativePath = currentPath.value ? currentPath.value + '/' + fileName : fileName
      const content = await App.ReadSkillFile(skillId, relativePath)
      fileContent.value = content
      selectedFile.value = fileName
      currentTab.value = 'preview'
    } catch (e) {
      error.value = String(e)
    } finally {
      loading.value = false
    }
  }
}

// 返回上一级目录
async function goUpDirectory() {
  const parts = currentPath.value.split('/')
  parts.pop()
  currentPath.value = parts.join('/')
  await loadFiles()
  selectedFile.value = null
  fileContent.value = ''
}

// 返回首页
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

// 使用 marked 渲染 Markdown
function renderMarkdown(content: string): string {
  return marked.parse(content) as string
}

// 获取当前选中的文件名
const selectedFileName = computed(() => {
  return selectedFile.value || ''
})

// 获取代码行数
const lineCount = computed(() => {
  return fileContent.value.split('\n').length
})

// 获取代码行数组
const fileLines = computed(() => fileContent.value.split('\n'))

// 获取当前选中的文件信息
const currentFileInfo = computed(() => {
  if (!selectedFile.value) return null
  return files.value.find(f => f.name === selectedFile.value)
})

// 面包屑路径
const breadcrumbParts = computed(() => {
  if (!currentPath.value) return []
  return currentPath.value.split('/')
})

// 复制内容到剪贴板
async function copyContent() {
  try {
    await navigator.clipboard.writeText(fileContent.value)
    console.log('复制成功')
  } catch (e) {
    console.error('复制失败:', e)
  }
}

// 下载文件
function downloadFile() {
  if (!selectedFile.value) return
  const blob = new Blob([fileContent.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = selectedFile.value
  a.click()
  URL.revokeObjectURL(url)
}
</script>

<template>
  <div class="h-screen flex flex-col bg-cyber-dark text-gray-300 font-sans">
    <!-- Header -->
    <header class="glass-panel border-b border-cyber-border">
      <div class="px-6 py-3">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <button @click="goBack" class="p-2 rounded-lg bg-cyber-panel border border-cyber-border hover:border-cyber-accent/50 transition-all">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
              </svg>
            </button>
            <div class="flex items-center gap-2">
              <span class="text-gray-500 text-sm">{{ skillId }}</span>
              <span class="text-gray-600">/</span>
              <span class="text-white font-mono text-sm">{{ selectedFileName || currentPath || '选择文件' }}</span>
            </div>
          </div>

          <div class="flex items-center gap-4">
            <!-- Tabs -->
            <div class="flex items-center gap-2">
              <button
                @click="currentTab = 'files'"
                :class="[
                  'px-4 py-2 rounded-lg border text-sm flex items-center gap-2 transition-all',
                  currentTab === 'files' ? 'bg-cyber-accent/20 border-cyber-accent/50 text-cyber-accent' : 'border-cyber-border'
                ]"
              >
                <i class="fas fa-folder-tree"></i>
                <span>文件</span>
              </button>
              <button
                @click="currentTab = 'preview'"
                :class="[
                  'px-4 py-2 rounded-lg border text-sm flex items-center gap-2 transition-all',
                  currentTab === 'preview' ? 'bg-cyber-accent/20 border-cyber-accent/50 text-cyber-accent' : 'border-cyber-border'
                ]"
              >
                <i class="fas fa-eye"></i>
                <span>预览</span>
              </button>
              <button
                @click="currentTab = 'code'"
                :class="[
                  'px-4 py-2 rounded-lg border text-sm flex items-center gap-2 transition-all',
                  currentTab === 'code' ? 'bg-cyber-accent/20 border-cyber-accent/50 text-cyber-accent' : 'border-cyber-border'
                ]"
              >
                <i class="fas fa-code"></i>
                <span>代码</span>
              </button>
            </div>

            <!-- 操作按钮 -->
            <div v-if="selectedFile" class="flex items-center gap-2">
              <button @click="copyContent" class="btn-icon" title="复制">
                <i class="fas fa-copy"></i>
              </button>
              <button @click="downloadFile" class="btn-icon" title="下载">
                <i class="fas fa-download"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="flex-1 overflow-hidden flex">
      <!-- File Tree Sidebar -->
      <aside class="w-64 border-r border-cyber-border bg-cyber-dark/50 overflow-y-auto flex-shrink-0">
        <div class="p-4">
          <h3 class="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-4">
            <i class="fas fa-folder-tree mr-2"></i>文件目录
          </h3>

          <!-- 面包屑导航 -->
          <div v-if="currentPath" class="mb-4 flex items-center gap-1 flex-wrap">
            <button
              @click="currentPath = ''; loadFiles()"
              class="text-xs text-cyber-accent hover:underline"
            >
              根目录
            </button>
            <template v-for="(part, index) in breadcrumbParts" :key="index">
              <span class="text-gray-600 text-xs">/</span>
              <button
                @click="currentPath = breadcrumbParts.slice(0, index + 1).join('/'); loadFiles()"
                class="text-xs text-cyber-accent hover:underline"
              >
                {{ part }}
              </button>
            </template>
          </div>

          <!-- 返回上一级 -->
          <button
            v-if="currentPath"
            @click="goUpDirectory"
            class="w-full mb-2 flex items-center gap-2 px-3 py-2 rounded-lg cursor-pointer transition-all text-sm hover:bg-cyber-accent/5 text-gray-400"
          >
            <i class="fas fa-level-up-alt w-4"></i>
            <span>..</span>
          </button>

          <!-- Loading -->
          <div v-if="loading" class="text-center py-8">
            <div class="w-8 h-8 mx-auto border-4 border-cyber-border border-t-cyber-accent rounded-full animate-spin"></div>
          </div>

          <!-- Error -->
          <div v-else-if="error" class="text-center py-4">
            <p class="text-red-400 text-sm">{{ error }}</p>
          </div>

          <!-- Empty State -->
          <div v-else-if="files.length === 0" class="text-center py-8">
            <i class="fas fa-folder-open text-3xl text-gray-600 mb-3"></i>
            <p class="text-sm text-gray-500">无文件</p>
          </div>

          <!-- File List -->
          <div v-else class="space-y-1">
            <div
              v-for="file in files"
              :key="file.path"
              @click="selectFile(file.name, file.is_dir)"
              :class="[
                'flex items-center gap-2 px-3 py-2 rounded-lg cursor-pointer transition-all text-sm',
                selectedFile === file.name
                  ? 'bg-cyber-accent/10 border-l-2 border-cyber-accent text-cyber-accent'
                  : 'hover:bg-cyber-accent/5 text-gray-400'
              ]"
            >
              <i :class="['fas', getFileIcon(file.name, file.is_dir)]" class="w-4"></i>
              <span class="truncate flex-1">{{ file.name }}</span>
              <i v-if="file.is_dir" class="fas fa-chevron-right text-xs text-gray-600"></i>
            </div>
          </div>
        </div>
      </aside>

      <!-- Content Panel -->
      <div class="flex-1 overflow-hidden bg-cyber-dark">
        <!-- Loading -->
        <div v-if="loading" class="h-full flex items-center justify-center">
          <div class="text-center">
            <div class="w-12 h-12 mx-auto border-4 border-cyber-border border-t-cyber-accent rounded-full animate-spin"></div>
            <p class="mt-4 text-gray-500">加载中...</p>
          </div>
        </div>

        <!-- Empty State -->
        <div v-else-if="!selectedFile" class="h-full flex items-center justify-center">
          <div class="text-center">
            <i class="fas fa-file-code text-5xl text-gray-600 mb-4"></i>
            <h3 class="text-lg font-semibold text-gray-400 mb-2">未选择文件</h3>
            <p class="text-sm text-gray-500">请从左侧目录选择要查看的文件</p>
          </div>
        </div>

        <!-- Preview View -->
        <div v-else-if="currentTab === 'preview'" class="h-full overflow-y-auto">
          <div class="max-w-4xl mx-auto p-8">
            <div class="markdown-body" v-html="renderMarkdown(fileContent)"></div>
          </div>
        </div>

        <!-- Code View -->
        <div v-else-if="currentTab === 'code'" class="h-full overflow-auto">
          <div class="p-6">
            <!-- File Header -->
            <div class="flex items-center justify-between mb-4 pb-4 border-b border-cyber-border">
              <div class="flex items-center gap-2">
                <i :class="['fas', getFileIcon(currentFileInfo?.name || '', false)]" class="text-gray-500"></i>
                <span class="text-sm text-gray-400">{{ currentFileInfo?.name }}</span>
              </div>
              <div class="text-xs text-gray-500">
                <span>{{ lineCount }} 行</span>
                <span class="mx-2">-</span>
                <span>{{ currentFileInfo?.size || 0 }} bytes</span>
              </div>
            </div>

            <!-- Code Content with Line Numbers -->
            <pre class="code-view font-mono text-sm">
              <code>
                <div v-for="(line, index) in fileLines" :key="index" class="flex">
                  <span class="line-number w-12 text-right pr-4 text-gray-600 select-none">{{ index + 1 }}</span>
                  <span class="flex-1 text-gray-300">{{ line }}</span>
                </div>
              </code>
            </pre>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
/* 图标按钮 */
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

/* Markdown Preview Styles */
.markdown-body {
  color: #e5e7eb;
  line-height: 1.8;
}

.markdown-body :deep(h1) {
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 1rem;
  color: #fff;
  border-bottom: 2px solid #1e1e2e;
  padding-bottom: 0.5rem;
}

.markdown-body :deep(h2) {
  font-size: 1.5rem;
  font-weight: 600;
  margin-top: 2rem;
  margin-bottom: 1rem;
  color: #00d4aa;
}

.markdown-body :deep(h3) {
  font-size: 1.25rem;
  font-weight: 600;
  margin-top: 1.5rem;
  margin-bottom: 0.75rem;
  color: #a855f7;
}

.markdown-body :deep(h4) {
  font-size: 1.1rem;
  font-weight: 600;
  margin-top: 1.25rem;
  margin-bottom: 0.5rem;
  color: #60a5fa;
}

.markdown-body :deep(p) {
  margin: 1rem 0;
  color: #d1d5db;
}

.markdown-body :deep(code) {
  background: #1e1e2e;
  padding: 0.2em 0.4em;
  border-radius: 4px;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 0.9em;
  color: #00d4aa;
}

.markdown-body :deep(pre) {
  background: #12121a;
  border: 1px solid #1e1e2e;
  border-radius: 8px;
  padding: 1rem;
  overflow-x: auto;
  margin: 1rem 0;
}

.markdown-body :deep(pre code) {
  background: transparent;
  padding: 0;
  color: #e5e7eb;
}

.markdown-body :deep(blockquote) {
  border-left: 4px solid #00d4aa;
  padding-left: 1rem;
  margin: 1rem 0;
  color: #9ca3af;
  font-style: italic;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin: 1rem 0;
  padding-left: 1.5rem;
}

.markdown-body :deep(li) {
  margin: 0.5rem 0;
  color: #d1d5db;
}

.markdown-body :deep(a) {
  color: #00d4aa;
  text-decoration: none;
}

.markdown-body :deep(a:hover) {
  text-decoration: underline;
}

.markdown-body :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1rem 0;
}

.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid #1e1e2e;
  padding: 0.75rem;
  text-align: left;
}

.markdown-body :deep(th) {
  background: #12121a;
  color: #fff;
  font-weight: 600;
}

.markdown-body :deep(hr) {
  border: none;
  border-top: 1px solid #1e1e2e;
  margin: 2rem 0;
}

.markdown-body :deep(img) {
  max-width: 100%;
  border-radius: 8px;
  margin: 1rem 0;
}

/* Code View Styles */
.code-view {
  background: #12121a;
  border: 1px solid #1e1e2e;
  border-radius: 8px;
  padding: 1rem;
  overflow-x: auto;
}

.line-number {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}
</style>
