import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as App from '../../wailsjs/go/backend/App'
import { skill } from '../../wailsjs/go/models'

export type Theme = 'dark' | 'light' | 'system'

export const useSettingsStore = defineStore('settings', () => {
  const theme = ref<Theme>('system')
  const autoUpdateCheck = ref(true)
  const updateFrequency = ref<'startup' | 'daily' | 'weekly'>('daily')
  const skillhubPath = ref('')
  const isFirstRun = ref(true)
  const detectedTools = ref<skill.ToolInfo[]>([])
  const isLoading = ref(false)
  const customCategories = ref<string[]>([]) // 自定义分类

  async function loadSettings() {
    isLoading.value = true
    try {
      const settings = await App.GetSettings()
      if (settings) {
        theme.value = (settings.theme as Theme) || 'system'
        autoUpdateCheck.value = settings.auto_update_check ?? true
        updateFrequency.value = (settings.update_frequency as 'startup' | 'daily' | 'weekly') || 'daily'
        isFirstRun.value = settings.first_run ?? true
        customCategories.value = settings.custom_categories || []
      }

      // Load SkillHub path
      skillhubPath.value = await App.GetSkillHubPath()

      // Load detected tools
      detectedTools.value = await App.GetDetectedTools()
    } catch (e) {
      console.error('Failed to load settings:', e)
    } finally {
      isLoading.value = false
    }
  }

  async function saveSettings() {
    try {
      const settings: skill.AppSettings = {
        skillhub_path: skillhubPath.value,
        theme: theme.value,
        auto_update_check: autoUpdateCheck.value,
        update_frequency: updateFrequency.value,
        first_run: isFirstRun.value,
        custom_categories: customCategories.value,
        github_token: ''
      }
      await App.SaveSettings(settings)
    } catch (e) {
      console.error('Failed to save settings:', e)
      throw e
    }
  }

  async function setSkillHubPath(path: string) {
    try {
      await App.SetSkillHubPath(path)
      skillhubPath.value = path
    } catch (e) {
      console.error('Failed to set SkillHub path:', e)
      throw e
    }
  }

  async function detectTools() {
    try {
      detectedTools.value = await App.GetDetectedTools()
    } catch (e) {
      console.error('Failed to detect tools:', e)
    }
  }

  function setTheme(newTheme: Theme) {
    theme.value = newTheme
    applyTheme(newTheme)
  }

  function applyTheme(themeValue: Theme) {
    // 移除所有主题类
    document.documentElement.classList.remove('dark-theme', 'light-theme')

    if (themeValue === 'system') {
      // 系统模式：使用 CSS 默认变量（暗色）
      // 不添加额外类，让浏览器自动适配
    } else if (themeValue === 'dark') {
      // 暗黑模式：添加 dark-theme 类
      document.documentElement.classList.add('dark-theme')
    } else if (themeValue === 'light') {
      // 亮色模式：添加 light-theme 类
      document.documentElement.classList.add('light-theme')
    }
  }

  async function completeFirstRun() {
    isFirstRun.value = false
    await saveSettings()
  }

  return {
    theme,
    autoUpdateCheck,
    updateFrequency,
    skillhubPath,
    isFirstRun,
    detectedTools,
    isLoading,
    customCategories,
    loadSettings,
    saveSettings,
    setSkillHubPath,
    detectTools,
    setTheme,
    applyTheme,
    completeFirstRun
  }
})
