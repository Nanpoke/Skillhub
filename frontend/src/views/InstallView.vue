<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import * as App from '../../wailsjs/go/backend/App'
import { skill } from '../../wailsjs/go/models'
import { useSkillStore } from '../stores/skills'
import { OnFileDrop, OnFileDropOff } from '../../wailsjs/runtime/runtime'

const router = useRouter()
const skillStore = useSkillStore()

// 标签相关
const allTags = computed(() => skillStore.allTags)

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

// 当前 Tab
const activeTab = ref<'local' | 'git'>('local')

// === 本地安装 Tab ===
const localLoading = ref(false)
const localError = ref<string | null>(null)
const localStep = ref<'drop' | 'scanning' | 'preview' | 'installing' | 'done'>('drop')
const selectedPath = ref('')
const localScanResult = ref<skill.LocalScanResult | null>(null)
const localTempPath = ref('')
const selectedLocalSkills = ref<Set<number>>(new Set())
const localSkillOptions = ref<Map<number, skill.InstallOptions>>(new Map())
const isDragover = ref(false)

// === Git 安装 Tab ===
const gitUrl = ref('')
const gitLoading = ref(false)
const gitLoadingStep = ref('')
const gitError = ref<string | null>(null)
const gitStep = ref<'input' | 'preview' | 'installing' | 'done'>('input')
const urlInfo = ref<any>(null)
const cloneResult = ref<skill.GitInstallResult | null>(null)
const tempPath = ref('')
const selectedSkills = ref<number[]>([])
const gitOptions = ref<Record<number, skill.InstallOptions>>({})
let loadingTimer: number | null = null

// 辅助函数：检查 Skill 是否已安装
function isSkillInstalled(skillInfo: any): boolean {
  const skillId = `${skillInfo.author}-${skillInfo.name}`
  return cloneResult.value?.InstalledSkills?.includes(skillId) || false
}

// 计算已安装的 Skill 数量
const installedCount = computed(() => {
  return cloneResult.value?.InstalledSkills?.length || 0
})

// 计算未安装的 Skill 数量
const availableCount = computed(() => {
  const total = cloneResult.value?.Skills?.length || 0
  return total - installedCount.value
})

const categories = computed(() => skillStore.categories)

// 处理 Wails 原生拖拽事件
function handleWailsFileDrop(x: number, y: number, paths: string[]) {
  // 只在本地安装 Tab 激活时处理
  if (activeTab.value !== 'local') return
  // 只在初始状态时处理
  if (localStep.value !== 'drop') return

  if (paths && paths.length > 0) {
    // 取第一个文件/文件夹路径
    selectedPath.value = paths[0]
    localError.value = null
    // 扫描路径
    scanLocalPath()
  }
}

onMounted(async () => {
  await skillStore.loadSkills()

  // 注册 Wails 文件拖拽监听
  OnFileDrop(handleWailsFileDrop, false)
})

onUnmounted(() => {
  // 取消注册 Wails 文件拖拽监听
  OnFileDropOff()
  // 清除加载定时器
  if (loadingTimer) {
    clearTimeout(loadingTimer)
  }
})

// === 本地安装功能 ===
function handleDragOver(e: DragEvent) {
  e.preventDefault()
  isDragover.value = true
}

function handleDragLeave() {
  isDragover.value = false
}

async function handleDrop(e: DragEvent) {
  e.preventDefault()
  isDragover.value = false
  // 在 Wails 中，拖拽无法获取完整文件路径，需要通过文件对话框选择
  // 显示提示让用户点击选择
}

// === 本地安装功能 ===
async function selectFolder() {
  try {
    const path = await App.SelectFolder()
    if (path) {
      selectedPath.value = path
      await scanLocalPath()
    }
  } catch (e) {
    localError.value = String(e)
  }
}

async function selectInstallFile() {
  try {
    const path = await App.SelectInstallFile()
    if (path) {
      selectedPath.value = path
      await scanLocalPath()
    }
  } catch (e) {
    localError.value = String(e)
  }
}

async function scanLocalPath() {
  if (!selectedPath.value) return

  localLoading.value = true
  localError.value = null
  localStep.value = 'scanning'

  try {
    const result = await App.ScanLocalPath(selectedPath.value)
    localScanResult.value = result
    localTempPath.value = result.TempPath || ''

    // 默认选中所有 Skills
    if (result.Skills) {
      result.Skills.forEach((_: any, index: number) => {
        selectedLocalSkills.value.add(index)
        localSkillOptions.value.set(index, {
          category: '',
          tags: [],
          notes: ''
        })
      })
    }

    localStep.value = 'preview'
  } catch (e) {
    localError.value = String(e)
    localStep.value = 'drop'
  } finally {
    localLoading.value = false
  }
}

function toggleLocalSkill(index: number) {
  if (selectedLocalSkills.value.has(index)) {
    selectedLocalSkills.value.delete(index)
  } else {
    selectedLocalSkills.value.add(index)
  }
}

function getLocalSkillOptions(index: number): skill.InstallOptions {
  if (!localSkillOptions.value.has(index)) {
    localSkillOptions.value.set(index, { category: '', tags: [], notes: '' })
  }
  return localSkillOptions.value.get(index)!
}

function addLocalSkillTag(index: number, event: Event) {
  const input = event.target as HTMLInputElement
  const tag = input.value.trim()
  if (tag && !getLocalSkillOptions(index).tags?.includes(tag)) {
    if (!getLocalSkillOptions(index).tags) {
      getLocalSkillOptions(index).tags = []
    }
    getLocalSkillOptions(index).tags!.push(tag)
  }
  input.value = ''
}

function removeLocalSkillTag(skillIndex: number, tagIndex: number) {
  const options = getLocalSkillOptions(skillIndex)
  if (options.tags) {
    options.tags.splice(tagIndex, 1)
  }
}

function clearLocalSelection() {
  selectedPath.value = ''
  localError.value = null
  localStep.value = 'drop'
  localScanResult.value = null
  localTempPath.value = ''
  selectedLocalSkills.value.clear()
  localSkillOptions.value.clear()
}

const selectedLocalCount = computed(() => selectedLocalSkills.value.size)

async function installLocalSkills() {
  if (selectedLocalSkills.value.size === 0) {
    localError.value = '请至少选择一个 Skill'
    return
  }

  localStep.value = 'installing'
  localError.value = null

  try {
    const skills = localScanResult.value?.Skills || []
    const tempPath = localTempPath.value
    const sourcePath = selectedPath.value

    for (const index of selectedLocalSkills.value) {
      const skillInfo = skills[index]
      const options = getLocalSkillOptions(index)
      const skillPath = skillInfo.path || ''
      await App.InstallFromLocalTemp(tempPath, skillPath, sourcePath, options)
    }

    // 清理临时目录
    if (tempPath) {
      App.CleanupLocalTemp(tempPath)
    }

    localStep.value = 'done'
    await skillStore.loadSkills()
  } catch (e) {
    localError.value = String(e)
    localStep.value = 'preview'
  }
}

// === Git 安装功能 ===
async function parseAndClone() {
  if (!gitUrl.value.trim()) {
    gitError.value = '请输入 Git 仓库 URL'
    return
  }

  gitLoading.value = true
  gitError.value = null
  gitLoadingStep.value = '正在连接Git仓库...'

  // 步骤切换定时器
  loadingTimer = window.setTimeout(() => {
    gitLoadingStep.value = '正在下载仓库内容...'
  }, 1500)
  loadingTimer = window.setTimeout(() => {
    gitLoadingStep.value = '正在识别可用Skill...'
  }, 3000)
  loadingTimer = window.setTimeout(() => {
    gitLoadingStep.value = '正在检查安装状态...'
  }, 4500)

  try {
    const info = await App.ParseGitURL(gitUrl.value)
    urlInfo.value = info
    // 不要在这里切换 gitStep，等 CloneFromGit 完成后再切换

    // 使用用户输入的完整URL进行克隆（包含子路径）
    const result = await App.CloneFromGit(gitUrl.value)
    cloneResult.value = result
    tempPath.value = result.TempPath

    // 克隆完成后再切换到预览状态
    gitStep.value = 'preview'

    // 清空之前的选中状态
    selectedSkills.value = []
    gitOptions.value = {}

    if (result.Skills) {
      result.Skills.forEach((_: any, index: number) => {
        // 只为安装选项创建默认配置，但不选中
        gitOptions.value[index] = {
          category: '',
          tags: [],
          notes: ''
        }
        // 不再默认选中所有 Skills
      })
    }
  } catch (e) {
    gitError.value = String(e)
    gitStep.value = 'input'
  } finally {
    gitLoading.value = false
    if (loadingTimer) {
      clearTimeout(loadingTimer)
      loadingTimer = null
    }
  }
}

function toggleSkill(index: number) {
  const pos = selectedSkills.value.indexOf(index)
  if (pos > -1) {
    selectedSkills.value.splice(pos, 1)
  } else {
    selectedSkills.value.push(index)
  }
}

function getGitOptions(index: number): skill.InstallOptions {
  if (!gitOptions.value[index]) {
    gitOptions.value[index] = { category: '', tags: [], notes: '' }
  }
  return gitOptions.value[index]
}

function addGitTag(index: number, event: Event) {
  const input = event.target as HTMLInputElement
  const tag = input.value.trim()
  if (tag && !getGitOptions(index).tags?.includes(tag)) {
    if (!getGitOptions(index).tags) {
      getGitOptions(index).tags = []
    }
    getGitOptions(index).tags!.push(tag)
  }
  input.value = ''
}

function removeGitTag(skillIndex: number, tagIndex: number) {
  const options = getGitOptions(skillIndex)
  if (options.tags) {
    options.tags.splice(tagIndex, 1)
  }
}

async function installFromGit() {
  if (selectedSkills.value.length === 0) {
    gitError.value = '请至少选择一个 Skill'
    return
  }

  gitStep.value = 'installing'
  gitError.value = null

  try {
    const skills = cloneResult.value?.Skills || []
    const gitUrl = urlInfo.value?.full_url || cloneResult.value?.GitURL || '' // 获取 Git URL

    for (const index of selectedSkills.value) {
      const skillInfo = skills[index]
      const options = getGitOptions(index)
      // 传递 Git URL
      await App.InstallFromGit(tempPath.value, skillInfo.path || '', gitUrl, options)
    }

    gitStep.value = 'done'
    App.CleanupClone(tempPath.value)
    tempPath.value = ''
    await skillStore.loadSkills()
  } catch (e) {
    gitError.value = String(e)
    gitStep.value = 'preview'
  }
}

function goBack() {
  router.push('/')
}

function goHome() {
  router.push('/')
}

const selectedCount = computed(() => selectedSkills.value.length)
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-cyber-dark via-cyber-panel to-cyber-dark text-gray-300 font-sans">
    <!-- Header -->
    <header class="glass-panel rounded-none sticky top-0 z-50 border-b border-cyber-border">
      <div class="max-w-5xl mx-auto px-6 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-4">
            <button @click="goBack" class="btn-icon">
              <i class="fas fa-arrow-left text-sm"></i>
            </button>
            <div>
              <h1 class="text-xl font-bold font-mono">
                <span class="gradient-text">安装</span> <span class="text-white">Skills</span>
              </h1>
            </div>
          </div>
          <button @click="goBack" class="btn-icon">
            <i class="fas fa-times text-sm"></i>
          </button>
        </div>
      </div>
    </header>

    <!-- Tab Navigation -->
    <div class="border-b border-cyber-border bg-cyber-panel/30">
      <div class="max-w-5xl mx-auto px-6">
        <div class="flex gap-2 py-3">
          <button
            @click="activeTab = 'local'"
            :class="[
              'flex items-center gap-3 px-6 py-3 rounded-[10px] transition-all',
              activeTab === 'local'
                ? 'bg-blue-500/10 border border-blue-500/30 text-blue-400'
                : 'bg-cyber-dark border border-cyber-border text-gray-400 hover:border-blue-500/30'
            ]"
          >
            <div :class="[
              'w-10 h-10 rounded-lg flex items-center justify-center',
              activeTab === 'local'
                ? 'bg-gradient-to-br from-blue-500/20 to-cyan-500/20'
                : 'bg-cyber-panel'
            ]">
              <i class="fas fa-folder-open text-lg"></i>
            </div>
            <div class="text-left">
              <div class="font-medium">本地安装</div>
              <div class="text-xs opacity-60">拖拽文件夹</div>
            </div>
          </button>

          <button
            @click="activeTab = 'git'"
            :class="[
              'flex items-center gap-3 px-6 py-3 rounded-[10px] transition-all',
              activeTab === 'git'
                ? 'bg-purple-500/10 border border-purple-500/30 text-purple-400'
                : 'bg-cyber-dark border border-cyber-border text-gray-400 hover:border-purple-500/30'
            ]"
          >
            <div :class="[
              'w-10 h-10 rounded-lg flex items-center justify-center',
              activeTab === 'git'
                ? 'bg-gradient-to-br from-purple-500/20 to-pink-500/20'
                : 'bg-cyber-panel'
            ]">
              <i class="fab fa-github text-lg"></i>
            </div>
            <div class="text-left">
              <div class="font-medium">Git 安装</div>
              <div class="text-xs opacity-60">从 GitHub 仓库</div>
            </div>
          </button>

          <!-- 浏览技能 - 外链 -->
          <a
            href="https://skills.sh"
            target="_blank"
            class="flex items-center gap-3 px-6 py-3 rounded-[10px] transition-all bg-cyber-dark border border-cyber-border text-gray-400 hover:border-cyber-accent/30 hover:text-cyber-accent"
          >
            <div class="w-10 h-10 rounded-lg flex items-center justify-center bg-cyber-panel">
              <i class="fas fa-globe text-lg"></i>
            </div>
            <div class="text-left">
              <div class="font-medium">浏览技能</div>
              <div class="text-xs opacity-60">访问 skills.sh</div>
            </div>
            <i class="fas fa-external-link-alt text-xs ml-auto"></i>
          </a>
        </div>
      </div>
    </div>

    <!-- Tab Content -->
    <main class="max-w-5xl mx-auto p-6">
      <!-- ========== 本地安装 Tab ========== -->
      <div v-if="activeTab === 'local'">
        <!-- Drop Zone -->
        <div v-if="localStep === 'drop'" class="space-y-6">
          <div
            class="glass-panel rounded-3xl p-8"
            @dragover="handleDragOver"
            @dragleave="handleDragLeave"
            @drop="handleDrop"
          >
            <div
              :class="[
                'rounded-2xl p-12 text-center transition-all',
                isDragover
                  ? 'border-2 border-cyber-accent bg-cyber-accent/10'
                  : 'border-2 border-dashed border-cyber-border'
              ]"
            >
              <div class="w-24 h-24 mx-auto mb-6 rounded-2xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 border border-blue-500/30 flex items-center justify-center">
                <i class="fas fa-cloud-upload-alt text-4xl text-blue-400"></i>
              </div>
              <h2 class="text-xl font-semibold text-white mb-3">拖拽文件到此处</h2>
              <p class="text-sm text-gray-500 mb-6">
                支持文件夹、<span class="text-cyber-accent">.zip</span> 或 <span class="text-cyber-accent">.skill</span> 压缩包
              </p>
              <p class="text-xs text-gray-600 mb-4">或</p>
              <div class="flex items-center justify-center gap-4">
                <button
                  @click="selectFolder"
                  class="px-6 py-3 rounded-xl bg-gradient-to-r from-blue-500 to-cyan-500 text-white font-medium hover:opacity-90 transition-all flex items-center gap-2"
                >
                  <i class="fas fa-folder-open"></i>
                  <span>选择文件夹</span>
                </button>
                <button
                  @click="selectInstallFile"
                  class="px-6 py-3 rounded-xl bg-cyber-panel border border-cyber-border text-gray-300 font-medium hover:border-cyber-accent/50 hover:text-cyber-accent transition-all flex items-center gap-2"
                >
                  <i class="fas fa-file-archive"></i>
                  <span>选择压缩包</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Supported Formats -->
          <div class="grid grid-cols-3 gap-4">
            <div class="glass-panel rounded-xl p-4 text-center">
              <div class="w-12 h-12 mx-auto mb-3 rounded-lg bg-cyber-panel flex items-center justify-center">
                <i class="fas fa-folder text-2xl text-yellow-400"></i>
              </div>
              <h4 class="text-sm font-medium text-white mb-1">文件夹</h4>
              <p class="text-xs text-gray-500">包含 SKILL.md 的目录</p>
            </div>

            <div class="glass-panel rounded-xl p-4 text-center">
              <div class="w-12 h-12 mx-auto mb-3 rounded-lg bg-cyber-panel flex items-center justify-center">
                <i class="fas fa-file-archive text-2xl text-purple-400"></i>
              </div>
              <h4 class="text-sm font-medium text-white mb-1">ZIP 压缩包</h4>
              <p class="text-xs text-gray-500">.zip 格式压缩文件</p>
            </div>

            <div class="glass-panel rounded-xl p-4 text-center">
              <div class="w-12 h-12 mx-auto mb-3 rounded-lg bg-cyber-panel flex items-center justify-center">
                <i class="fas fa-box text-2xl text-cyber-accent"></i>
              </div>
              <h4 class="text-sm font-medium text-white mb-1">Skill 包</h4>
              <p class="text-xs text-gray-500">.skill 专用格式</p>
            </div>
          </div>
        </div>

        <!-- Scanning -->
        <div v-else-if="localStep === 'scanning'" class="text-center py-12">
          <div class="w-20 h-20 mx-auto mb-6 rounded-full border-4 border-cyber-border border-t-cyber-accent animate-spin"></div>
          <h3 class="text-lg font-semibold text-white mb-2">正在扫描...</h3>
          <p class="text-sm text-gray-500">正在分析文件结构</p>
        </div>

        <!-- Preview State -->
        <div v-else-if="localStep === 'preview' && localScanResult" class="space-y-6">
          <!-- Source Info -->
          <div class="glass-panel rounded-2xl p-6 border border-cyber-border">
            <div class="flex items-center gap-4 mb-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/20 flex items-center justify-center">
                <i class="fas fa-folder-open text-xl text-blue-400" v-if="!localScanResult.IsZip"></i>
                <i class="fas fa-file-archive text-xl text-purple-400" v-else></i>
              </div>
              <div>
                <h3 class="font-semibold text-white">{{ selectedPath.split(/[/\\]/).pop() }}</h3>
                <p class="text-sm text-gray-500">发现 {{ localScanResult.Skills?.length || 0 }} 个 Skills</p>
              </div>
            </div>
          </div>

          <!-- Skill List -->
          <div class="space-y-4">
            <div
              v-for="(skillInfo, index) in localScanResult.Skills"
              :key="index"
              class="glass-panel rounded-2xl p-5 border border-cyber-border"
              :class="{ 'border-blue-500/50': selectedLocalSkills.has(index) }"
            >
              <div class="flex items-start gap-4">
                <button
                  @click="toggleLocalSkill(index)"
                  :class="[
                    'w-6 h-6 rounded-lg border-2 flex items-center justify-center transition-all flex-shrink-0 mt-1',
                    selectedLocalSkills.has(index)
                      ? 'bg-blue-500 border-blue-500'
                      : 'border-gray-500 hover:border-blue-500'
                  ]"
                >
                  <i v-if="selectedLocalSkills.has(index)" class="fas fa-check text-white text-xs"></i>
                </button>

                <div class="flex-1">
                  <h4 class="font-semibold text-white font-mono">{{ skillInfo.name }}</h4>
                  <p class="text-sm text-gray-500 mb-2">{{ skillInfo.author }}</p>
                  <p class="text-sm text-gray-400">{{ skillInfo.description || '暂无描述' }}</p>

                  <div v-if="selectedLocalSkills.has(index)" class="mt-4 pt-4 border-t border-cyber-border space-y-3">
                    <!-- 分类和标签在同一行 -->
                    <div class="grid grid-cols-2 gap-3">
                      <div>
                        <label class="text-xs text-gray-500 block mb-1">分类</label>
                        <select
                          v-model="getLocalSkillOptions(index).category"
                          class="w-full bg-cyber-dark border border-cyber-border rounded-lg py-2 px-3 text-sm focus:outline-none focus:border-blue-500"
                        >
                          <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
                        </select>
                      </div>
                      <div>
                        <label class="text-xs text-gray-500 block mb-1">标签</label>
                        <div class="flex flex-wrap gap-1 mb-1" v-if="getLocalSkillOptions(index).tags && getLocalSkillOptions(index).tags!.length > 0">
                          <span
                            v-for="(tag, tagIndex) in getLocalSkillOptions(index).tags"
                            :key="tagIndex"
                            class="tag px-2 py-0.5 rounded-lg text-xs text-blue-400 flex items-center gap-1"
                          >
                            {{ tag }}
                            <i @click="removeLocalSkillTag(index, tagIndex)" class="fas fa-times cursor-pointer hover:text-red-400"></i>
                          </span>
                        </div>
                        <!-- 标签输入框带自动补全 -->
                        <div class="relative mb-2">
                          <input
                            @keydown.enter.prevent="addLocalSkillTag(index, $event)"
                            type="text"
                            placeholder="输入标签后按回车"
                            list="local-tags-list"
                            class="w-full bg-cyber-dark border border-cyber-border rounded-lg py-2 px-3 text-sm focus:outline-none focus:border-blue-500"
                          />
                          <datalist id="local-tags-list">
                            <option v-for="tag in allTags" :key="tag" :value="tag">{{ tag }}</option>
                          </datalist>
                        </div>
                        <!-- 常用标签选择 -->
                        <div v-if="popularTags.length > 0" class="mb-1">
                          <p class="text-xs text-gray-500 mb-1">常用标签：</p>
                          <div class="flex flex-wrap gap-1">
                            <button
                              v-for="tag in popularTags"
                              :key="tag"
                              @click="() => {
                                const options = getLocalSkillOptions(index)
                                if (!options.tags) options.tags = []
                                if (!options.tags.includes(tag)) {
                                  options.tags.push(tag)
                                }
                              }"
                              :disabled="getLocalSkillOptions(index).tags?.includes(tag)"
                              :class="[
                                'px-2 py-1 rounded-lg text-xs transition-all flex items-center gap-1',
                                getLocalSkillOptions(index).tags?.includes(tag)
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
                    <div>
                      <label class="text-xs text-gray-500 block mb-1">备注</label>
                      <input
                        v-model="getLocalSkillOptions(index).notes"
                        type="text"
                        placeholder="添加备注..."
                        class="w-full bg-cyber-dark border border-cyber-border rounded-lg py-2 px-3 text-sm focus:outline-none focus:border-blue-500"
                      />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Error -->
          <div v-if="localError" class="p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
            {{ localError }}
          </div>

          <!-- Actions -->
          <div class="flex gap-4">
            <button
              @click="clearLocalSelection"
              class="flex-1 py-3 rounded-xl bg-cyber-panel border border-cyber-border text-gray-300 font-medium hover:border-blue-500/50 transition-all"
            >
              取消
            </button>
            <button
              @click="installLocalSkills"
              :disabled="selectedLocalCount === 0"
              class="flex-1 py-3 rounded-xl bg-gradient-to-r from-blue-500 to-cyan-500 text-white font-semibold hover:opacity-90 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              安装 {{ selectedLocalCount }} 个 Skills
            </button>
          </div>
        </div>

        <!-- Installing Progress -->
        <div v-else-if="localStep === 'installing'" class="text-center py-12">
          <div class="w-20 h-20 mx-auto mb-6 rounded-full border-4 border-cyber-border border-t-cyber-accent animate-spin"></div>
          <h3 class="text-lg font-semibold text-white mb-2">正在安装...</h3>
          <p class="text-sm text-gray-500">正在处理文件，请稍候</p>
        </div>

        <!-- Success -->
        <div v-else-if="localStep === 'done'" class="text-center py-12">
          <div class="w-20 h-20 mx-auto mb-6 rounded-full bg-blue-500/20 flex items-center justify-center">
            <i class="fas fa-check text-4xl text-blue-400"></i>
          </div>
          <h3 class="text-xl font-semibold text-white mb-2">安装完成！</h3>
          <p class="text-sm text-gray-500 mb-6">已成功安装 {{ selectedLocalCount }} 个 Skills</p>
          <button
            @click="goHome"
            class="px-8 py-3 rounded-xl bg-gradient-to-r from-blue-500 to-cyan-500 text-white font-semibold hover:opacity-90 transition-all"
          >
            返回主页
          </button>
        </div>
      </div>

      <!-- ========== Git 安装 Tab ========== -->
      <div v-if="activeTab === 'git'">
        <!-- Step 1: Input URL -->
        <div v-if="gitStep === 'input'" class="space-y-6">
          <!-- 加载状态 -->
          <div v-if="gitLoading" class="text-center py-20">
            <div class="w-16 h-16 mx-auto mb-6 border-4 border-cyber-border border-t-purple-500 rounded-full animate-spin"></div>
            <p class="text-gray-300 text-lg mb-2">
              <i class="fas fa-circle-notch fa-spin mr-2 text-purple-400"></i>
              {{ gitLoadingStep }}
            </p>
            <p class="text-gray-500 text-sm">请稍候，这可能需要几秒钟时间</p>
          </div>

          <!-- 输入表单 -->
          <div v-else class="glass-panel rounded-2xl p-6 border border-cyber-border">
            <h3 class="text-lg font-semibold text-white mb-4">输入 Git 仓库地址</h3>

            <div class="space-y-4">
              <div>
                <label class="block text-sm text-gray-400 mb-2">仓库 URL</label>
                <input
                  v-model="gitUrl"
                  type="text"
                  placeholder="https://github.com/owner/repo 或 owner/repo"
                  class="w-full bg-cyber-dark border border-cyber-border rounded-xl py-3 px-4 text-sm focus:outline-none focus:border-purple-500 transition-all font-mono"
                  @keyup.enter="parseAndClone"
                />
              </div>

              <div class="text-xs text-gray-500">
                <p class="mb-2">支持的格式：</p>
                <ul class="list-disc list-inside space-y-1">
                  <li>https://github.com/owner/repo</li>
                  <li>owner/repo</li>
                  <li>https://github.com/owner/repo/tree/main/subdir</li>
                </ul>
              </div>

              <div v-if="gitError" class="p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
                {{ gitError }}
              </div>

              <button
                @click="parseAndClone"
                :disabled="gitLoading || !gitUrl.trim()"
                class="w-full py-3 rounded-xl bg-gradient-to-r from-purple-500 to-pink-500 text-white font-semibold hover:opacity-90 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
              >
                <div v-if="gitLoading" class="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                <span v-else><i class="fas fa-download mr-2"></i>克隆并扫描</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Step 2: Preview and Select -->
        <div v-if="gitStep === 'preview' && cloneResult" class="space-y-6">
          <div class="glass-panel rounded-2xl p-6 border border-cyber-border">
            <div class="flex items-center gap-4 mb-4">
              <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500/20 to-pink-500/20 flex items-center justify-center">
                <i class="fab fa-github text-xl text-purple-400"></i>
              </div>
              <div>
                <h3 class="font-semibold text-white">{{ urlInfo?.short_ref || gitUrl }}</h3>
                <p class="text-sm text-gray-500">
                  发现 {{ cloneResult.Skills?.length || 0 }} 个 Skills，其中 {{ installedCount }} 个已安装
                </p>
                <p v-if="availableCount > 0" class="text-xs text-purple-400 mt-1">
                  请选择 {{ availableCount }} 个未安装的 Skills 进行安装
                </p>
                <p v-else class="text-xs text-yellow-500 mt-1">
                  所有 Skills 已安装
                </p>
              </div>
            </div>
          </div>

          <div class="space-y-4">
            <div
              v-for="(skillInfo, index) in cloneResult.Skills"
              :key="index"
              class="glass-panel rounded-2xl p-5 border border-cyber-border"
              :class="{
                'border-purple-500/50': selectedSkills.includes(index),
                'opacity-50': isSkillInstalled(skillInfo)
              }"
            >
              <div class="flex items-start gap-4">
                <button
                  @click="toggleSkill(index)"
                  :disabled="isSkillInstalled(skillInfo)"
                  :class="[
                    'w-6 h-6 rounded-lg border-2 flex items-center justify-center transition-all flex-shrink-0 mt-1',
                    selectedSkills.includes(index)
                      ? 'bg-purple-500 border-purple-500'
                      : 'border-gray-500 hover:border-purple-500',
                    isSkillInstalled(skillInfo) ? 'cursor-not-allowed bg-gray-700 border-gray-600' : ''
                  ]"
                >
                  <i v-if="selectedSkills.includes(index)" class="fas fa-check text-white text-xs"></i>
                </button>

                <div class="flex-1">
                  <div class="flex items-center gap-2 mb-1">
                    <h4 class="font-semibold text-white font-mono">{{ skillInfo.name }}</h4>
                    <span v-if="isSkillInstalled(skillInfo)" class="px-2 py-0.5 rounded-full bg-green-500/20 text-green-400 text-xs border border-green-500/30">
                      已安装
                    </span>
                  </div>
                  <p class="text-sm text-gray-500 mb-2">{{ skillInfo.author }}</p>
                  <p class="text-sm text-gray-400">{{ skillInfo.description || '暂无描述' }}</p>

                  <div v-if="selectedSkills.includes(index) && !isSkillInstalled(skillInfo)" class="mt-4 pt-4 border-t border-cyber-border space-y-3">
                    <!-- 分类和标签在同一行 -->
                    <div class="grid grid-cols-2 gap-3">
                      <div>
                        <label class="text-xs text-gray-500 block mb-1">分类</label>
                        <select
                          v-model="getGitOptions(index).category"
                          class="w-full bg-cyber-dark border border-cyber-border rounded-lg py-2 px-3 text-sm focus:outline-none focus:border-purple-500"
                        >
                          <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
                        </select>
                      </div>
                      <div>
                        <label class="text-xs text-gray-500 block mb-1">标签</label>
                        <div class="flex flex-wrap gap-1 mb-1" v-if="getGitOptions(index).tags && getGitOptions(index).tags!.length > 0">
                          <span
                            v-for="(tag, tagIndex) in getGitOptions(index).tags"
                            :key="tagIndex"
                            class="tag px-2 py-0.5 rounded-lg text-xs text-purple-400 flex items-center gap-1"
                          >
                            {{ tag }}
                            <i @click="removeGitTag(index, tagIndex)" class="fas fa-times cursor-pointer hover:text-red-400"></i>
                          </span>
                        </div>
                        <!-- 标签输入框带自动补全 -->
                        <div class="relative mb-2">
                          <input
                            @keydown.enter.prevent="addGitTag(index, $event)"
                            type="text"
                            placeholder="输入标签后按回车"
                            list="git-tags-list"
                            class="w-full bg-cyber-dark border border-cyber-border rounded-lg py-2 px-3 text-sm focus:outline-none focus:border-purple-500"
                          />
                          <datalist id="git-tags-list">
                            <option v-for="tag in allTags" :key="tag" :value="tag">{{ tag }}</option>
                          </datalist>
                        </div>
                        <!-- 常用标签选择 -->
                        <div v-if="popularTags.length > 0" class="mb-1">
                          <p class="text-xs text-gray-500 mb-1">常用标签：</p>
                          <div class="flex flex-wrap gap-1">
                            <button
                              v-for="tag in popularTags"
                              :key="tag"
                              @click="() => {
                                const options = getGitOptions(index)
                                if (!options.tags) options.tags = []
                                if (!options.tags.includes(tag)) {
                                  options.tags.push(tag)
                                }
                              }"
                              :disabled="getGitOptions(index).tags?.includes(tag)"
                              :class="[
                                'px-2 py-1 rounded-lg text-xs transition-all flex items-center gap-1',
                                getGitOptions(index).tags?.includes(tag)
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
                    <div>
                      <label class="text-xs text-gray-500 block mb-1">备注</label>
                      <input
                        v-model="getGitOptions(index).notes"
                        type="text"
                        placeholder="添加备注..."
                        class="w-full bg-cyber-dark border border-cyber-border rounded-lg py-2 px-3 text-sm focus:outline-none focus:border-purple-500"
                      />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div v-if="gitError" class="p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
            {{ gitError }}
          </div>

          <div class="flex gap-4">
            <button
              @click="gitStep = 'input'; gitError = null"
              class="flex-1 py-3 rounded-xl bg-cyber-panel border border-cyber-border text-gray-300 font-medium hover:border-purple-500/50 transition-all"
            >
              取消
            </button>
            <button
              @click="installFromGit"
              :disabled="selectedCount === 0"
              class="flex-1 py-3 rounded-xl bg-gradient-to-r from-purple-500 to-pink-500 text-white font-semibold hover:opacity-90 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              安装 {{ selectedCount }} 个 Skills
            </button>
          </div>
        </div>

        <!-- Step 3: Installing -->
        <div v-if="gitStep === 'installing'" class="text-center py-12">
          <div class="w-16 h-16 mx-auto mb-6 border-4 border-cyber-border border-t-purple-500 rounded-full animate-spin"></div>
          <h3 class="text-lg font-semibold text-white mb-2">正在安装...</h3>
          <p class="text-sm text-gray-500">请稍候</p>
        </div>

        <!-- Step 4: Done -->
        <div v-if="gitStep === 'done'" class="text-center py-12">
          <div class="w-20 h-20 mx-auto mb-6 rounded-full bg-purple-500/20 flex items-center justify-center">
            <i class="fas fa-check text-4xl text-purple-400"></i>
          </div>
          <h3 class="text-xl font-semibold text-white mb-2">安装完成！</h3>
          <p class="text-sm text-gray-500 mb-6">已成功安装 {{ selectedCount }} 个 Skills</p>
          <button
            @click="goHome"
            class="px-8 py-3 rounded-xl bg-gradient-to-r from-purple-500 to-pink-500 text-white font-semibold hover:opacity-90 transition-all"
          >
            返回主页
          </button>
        </div>
      </div>
    </main>
  </div>
</template>
