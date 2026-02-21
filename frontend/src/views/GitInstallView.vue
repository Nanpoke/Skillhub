<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import * as App from '../../wailsjs/go/backend/App'
import { skill } from '../../wailsjs/go/models'

const router = useRouter()

const gitUrl = ref('')
const isLoading = ref(false)
const error = ref<string | null>(null)
const step = ref<'input' | 'preview' | 'installing' | 'done'>('input')

// 解析后的 URL 信息
const urlInfo = ref<GitURLInfo | null>(null)

// 克隆结果
const cloneResult = ref<skill.GitInstallResult | null>(null)
const tempPath = ref('')

// 选中的 Skills
const selectedSkills = ref<Set<number>>(new Set())

// 安装选项
const installOptions = ref<Map<number, skill.InstallOptions>>(new Map())

interface GitURLInfo {
  owner: string
  repo: string
  sub_path: string
  full_url: string
  short_ref: string
}

function goBack() {
  // 清理临时目录
  if (tempPath.value) {
    App.CleanupClone(tempPath.value)
  }
  router.push('/install')
}

async function parseAndClone() {
  if (!gitUrl.value.trim()) {
    error.value = '请输入 Git 仓库 URL'
    return
  }

  isLoading.value = true
  error.value = null

  try {
    // 解析 URL
    const info = await App.ParseGitURL(gitUrl.value)
    urlInfo.value = info

    step.value = 'preview'

    // 使用解析后的基础 URL 进行克隆
    const result = await App.CloneFromGit(info.full_url)
    cloneResult.value = result
    tempPath.value = result.TempPath

    // 初始化选中状态和选项
    if (result.Skills) {
      result.Skills.forEach((_: any, index: number) => {
        selectedSkills.value.add(index)
        installOptions.value.set(index, {
          category: '',
          tags: [],
          notes: ''
        })
      })
    }
  } catch (e) {
    error.value = String(e)
    step.value = 'input'
  } finally {
    isLoading.value = false
  }
}

function toggleSkill(index: number) {
  if (selectedSkills.value.has(index)) {
    selectedSkills.value.delete(index)
  } else {
    selectedSkills.value.add(index)
  }
}

function getSkillOptions(index: number): skill.InstallOptions {
  if (!installOptions.value.has(index)) {
    installOptions.value.set(index, {
      category: '',
      tags: [],
      notes: ''
    })
  }
  return installOptions.value.get(index)!
}

function addTag(index: number, event: Event) {
  const input = event.target as HTMLInputElement
  const tag = input.value.trim()
  if (tag && !getSkillOptions(index).tags?.includes(tag)) {
    if (!getSkillOptions(index).tags) {
      getSkillOptions(index).tags = []
    }
    getSkillOptions(index).tags!.push(tag)
  }
  input.value = ''
}

function removeTag(skillIndex: number, tagIndex: number) {
  const options = getSkillOptions(skillIndex)
  if (options.tags) {
    options.tags.splice(tagIndex, 1)
  }
}

async function installSelected() {
  if (selectedSkills.value.size === 0) {
    error.value = '请至少选择一个 Skill'
    return
  }

  step.value = 'installing'
  error.value = null

  try {
    const skills = cloneResult.value?.Skills || []
    const gitUrl = urlInfo.value?.full_url || cloneResult.value?.GitURL || '' // 获取 Git URL

    for (const index of selectedSkills.value) {
      const skillInfo = skills[index]
      const options = getSkillOptions(index)

      // 传递 Git URL
      await App.InstallFromGit(tempPath.value, skillInfo.path || '', gitUrl, options)
    }

    step.value = 'done'

    // 清理临时目录
    App.CleanupClone(tempPath.value)
    tempPath.value = ''
  } catch (e) {
    error.value = String(e)
    step.value = 'preview'
  }
}

function goHome() {
  router.push('/')
}

const selectedCount = computed(() => selectedSkills.value.size)

const categories = ['内容创作', '认知增强', '开发辅助', '数据分析', '教育学习', 'AI/LLM', '其他']
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
              <span class="gradient-text">Git</span> <span class="text-white">安装</span>
            </h1>
            <p class="text-xs text-gray-500">从 GitHub 仓库安装 Skill</p>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-4xl mx-auto p-6">
      <!-- Step 1: Input URL -->
      <div v-if="step === 'input'" class="space-y-6">
        <div class="glass-panel rounded-2xl p-6 border border-cyber-border">
          <h2 class="text-lg font-semibold text-white mb-4">输入 Git 仓库地址</h2>

          <div class="space-y-4">
            <div>
              <label class="block text-sm text-gray-400 mb-2">仓库 URL</label>
              <input
                v-model="gitUrl"
                type="text"
                placeholder="https://github.com/owner/repo 或 owner/repo"
                class="w-full bg-cyber-dark border border-cyber-border rounded-xl py-3 px-4 text-sm focus:outline-none focus:border-cyber-accent transition-all font-mono"
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

            <div v-if="error" class="p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
              {{ error }}
            </div>

            <button
              @click="parseAndClone"
              :disabled="isLoading || !gitUrl.trim()"
              class="w-full py-3 rounded-xl bg-gradient-to-r from-cyber-accent to-cyber-accent2 text-cyber-dark font-semibold hover:opacity-90 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              <div v-if="isLoading" class="w-5 h-5 border-2 border-cyber-dark/30 border-t-cyber-dark rounded-full animate-spin"></div>
              <span v-else>克隆并扫描</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Step 2: Preview and Select -->
      <div v-if="step === 'preview' && cloneResult" class="space-y-6">
        <!-- Repository Info -->
        <div class="glass-panel rounded-2xl p-6 border border-cyber-border">
          <div class="flex items-center gap-4 mb-4">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500/20 to-pink-500/20 flex items-center justify-center">
              <span class="text-xl">📦</span>
            </div>
            <div>
              <h3 class="font-semibold text-white">{{ urlInfo?.short_ref || gitUrl }}</h3>
              <p class="text-sm text-gray-500">发现 {{ cloneResult.Skills?.length || 0 }} 个 Skills</p>
            </div>
          </div>
        </div>

        <!-- Skills List -->
        <div class="space-y-4">
          <div
            v-for="(skillInfo, index) in cloneResult.Skills"
            :key="index"
            class="glass-panel rounded-2xl p-5 border border-cyber-border"
            :class="{ 'border-cyber-accent/50': selectedSkills.has(index) }"
          >
            <div class="flex items-start gap-4">
              <!-- Checkbox -->
              <button
                @click="toggleSkill(index)"
                :class="[
                  'w-6 h-6 rounded-lg border-2 flex items-center justify-center transition-all flex-shrink-0 mt-1',
                  selectedSkills.has(index)
                    ? 'bg-cyber-accent border-cyber-accent'
                    : 'border-gray-500 hover:border-cyber-accent'
                ]"
              >
                <svg v-if="selectedSkills.has(index)" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-cyber-dark" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                </svg>
              </button>

              <!-- Content -->
              <div class="flex-1">
                <h4 class="font-semibold text-white font-mono">{{ skillInfo.name }}</h4>
                <p class="text-sm text-gray-500 mb-2">{{ skillInfo.author }}</p>
                <p class="text-sm text-gray-400">{{ skillInfo.description || '暂无描述' }}</p>

                <!-- Install Options -->
                <div v-if="selectedSkills.has(index)" class="mt-4 pt-4 border-t border-cyber-border space-y-3">
                  <div>
                    <label class="text-xs text-gray-500 block mb-1">分类</label>
                    <select
                      v-model="getSkillOptions(index).category"
                      class="w-full bg-cyber-dark border border-cyber-border rounded-lg py-2 px-3 text-sm focus:outline-none focus:border-cyber-accent"
                    >
                      <option value="">选择分类</option>
                      <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
                    </select>
                  </div>
                  <div>
                    <label class="text-xs text-gray-500 block mb-1">标签</label>
                    <div class="flex flex-wrap gap-2 mb-2">
                      <span
                        v-for="(tag, tagIndex) in getSkillOptions(index).tags"
                        :key="tagIndex"
                        class="tag px-2 py-1 rounded-lg text-xs text-cyber-accent flex items-center gap-1"
                      >
                        {{ tag }}
                        <button @click="removeTag(index, tagIndex)" class="hover:text-red-400">×</button>
                      </span>
                    </div>
                    <input
                      @keydown.enter.prevent="addTag(index, $event)"
                      type="text"
                      placeholder="输入标签后按回车添加..."
                      class="w-full bg-cyber-dark border border-cyber-border rounded-lg py-2 px-3 text-sm focus:outline-none focus:border-cyber-accent"
                    />
                  </div>
                  <div>
                    <label class="text-xs text-gray-500 block mb-1">备注</label>
                    <input
                      v-model="getSkillOptions(index).notes"
                      type="text"
                      placeholder="添加备注..."
                      class="w-full bg-cyber-dark border border-cyber-border rounded-lg py-2 px-3 text-sm focus:outline-none focus:border-cyber-accent"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Error -->
        <div v-if="error" class="p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm">
          {{ error }}
        </div>

        <!-- Actions -->
        <div class="flex gap-4">
          <button
            @click="goBack"
            class="flex-1 py-3 rounded-xl bg-cyber-panel border border-cyber-border text-gray-300 font-medium hover:border-cyber-accent/50 transition-all"
          >
            取消
          </button>
          <button
            @click="installSelected"
            :disabled="selectedCount === 0"
            class="flex-1 py-3 rounded-xl bg-gradient-to-r from-cyber-accent to-cyber-accent2 text-cyber-dark font-semibold hover:opacity-90 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            安装 {{ selectedCount }} 个 Skills
          </button>
        </div>
      </div>

      <!-- Step 3: Installing -->
      <div v-if="step === 'installing'" class="text-center py-12">
        <div class="w-16 h-16 mx-auto mb-6 border-4 border-cyber-border border-t-cyber-accent rounded-full animate-spin"></div>
        <h3 class="text-lg font-semibold text-white mb-2">正在安装...</h3>
        <p class="text-sm text-gray-500">请稍候</p>
      </div>

      <!-- Step 4: Done -->
      <div v-if="step === 'done'" class="text-center py-12">
        <div class="w-20 h-20 mx-auto mb-6 rounded-full bg-cyber-accent/20 flex items-center justify-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-cyber-accent" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <h3 class="text-xl font-semibold text-white mb-2">安装完成！</h3>
        <p class="text-sm text-gray-500 mb-6">已成功安装 {{ selectedCount }} 个 Skills</p>
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
