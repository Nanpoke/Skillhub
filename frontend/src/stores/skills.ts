import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as App from '../../wailsjs/go/backend/App'
import { skill } from '../../wailsjs/go/models'

// 分类信息接口
interface CategoryInfo {
  name: string
  is_preset: boolean
}

// 预设分类
const presetCategories = ['全部', '内容创作', '认知增强', '开发辅助', '数据分析', '教育学习', 'AI/LLM', '其他']

export const useSkillStore = defineStore('skills', () => {
  const skills = ref<skill.Skill[]>([])
  const tools = ref<skill.ToolInfo[]>([])
  const customTools = ref<skill.ToolInfo[]>([]) // 自定义工具列表
  const customCategories = ref<string[]>([]) // 自定义分类
  const selectedCategory = ref('全部')
  const selectedTags = ref<string[]>([])
  const selectedTools = ref<string[]>([])  // 工具筛选状态
  const searchQuery = ref('')
  const isLoading = ref(false)
  const error = ref<string | null>(null)

  // 合并后的分类列表（用于侧边栏显示）
  const categories = computed(() => [
    ...presetCategories,
    ...customCategories.value.filter(c => !presetCategories.includes(c))
  ])

  // 获取分类详细信息（含是否预设）
  const allCategoriesWithInfo = computed<CategoryInfo[]>(() => {
    const result: CategoryInfo[] = []
    // 预设分类（排除"全部"）
    for (let i = 1; i < presetCategories.length; i++) {
      result.push({ name: presetCategories[i], is_preset: true })
    }
    // 自定义分类
    for (const name of customCategories.value) {
      if (!presetCategories.includes(name)) {
        result.push({ name, is_preset: false })
      }
    }
    return result
  })

  // Computed: Get all unique tags from skills
  const allTags = computed(() => {
    const tagSet = new Set<string>()
    skills.value.forEach(s => {
      s.tags?.forEach(t => tagSet.add(t))
    })
    return Array.from(tagSet).sort()
  })

  // Computed: Filter skills by category, tags, and search
  const filteredSkills = computed(() => {
    let result = skills.value

    // Filter by category
    if (selectedCategory.value !== '全部') {
      result = result.filter(s => s.category === selectedCategory.value)
    }

    // Filter by tags
    if (selectedTags.value.length > 0) {
      result = result.filter(s =>
        selectedTags.value.every(tag => s.tags?.includes(tag))
      )
    }

    // Filter by search query
    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase()
      result = result.filter(s =>
        s.name.toLowerCase().includes(query) ||
        s.description?.toLowerCase().includes(query) ||
        s.tags?.some(t => t.toLowerCase().includes(query)) ||
        s.author?.toLowerCase().includes(query)
      )
    }

    // Filter by tools (OR logic: show skills enabled in any selected tool)
    if (selectedTools.value.length > 0) {
      result = result.filter(s =>
        selectedTools.value.some(toolId => s.tools_enabled?.[toolId])
      )
    }

    return result
  })

  const totalSkills = computed(() => skills.value.length)

  const enabledSkills = computed(() =>
    skills.value.filter(s =>
      s.tools_enabled && Object.values(s.tools_enabled).some(v => v)
    ).length
  )

  // 合并后的已启用工具列表（预置 + 自定义）
  const allEnabledTools = computed(() => {
    const enabledBuiltIn = tools.value.filter(t => t.is_enabled)
    const enabledCustom = customTools.value.filter(t => t.is_enabled)
    return [...enabledBuiltIn, ...enabledCustom]
  })

  // Actions
  async function loadSkills() {
    isLoading.value = true
    error.value = null
    try {
      const result = await App.GetSkills()
      skills.value = result || []
    } catch (e) {
      console.error('Failed to load skills:', e)
      error.value = String(e)
    } finally {
      isLoading.value = false
    }
  }

  async function loadTools() {
    try {
      const result = await App.GetDetectedTools()
      tools.value = result || []
    } catch (e) {
      console.error('Failed to load tools:', e)
    }
  }

  // 新增：加载自定义工具
  async function loadCustomTools() {
    try {
      const custom = await App.GetCustomTools()
      if (custom) {
        // 转换为 ToolInfo 格式，添加类型断言
        customTools.value = custom.map((ct): skill.ToolInfo => ({
          id: ct.id,
          name: ct.name,
          skills_path: ct.skills_path,
          is_installed: true,
          is_enabled: ct.enabled
        }))
      } else {
        customTools.value = []
      }
    } catch (e) {
      console.error('Failed to load custom tools:', e)
    }
  }

  async function toggleTool(skillId: string, toolId: string) {
    try {
      const newState = await App.ToggleSkill(skillId, toolId)
      // Update local state
      const skill = skills.value.find(s => s.id === skillId)
      if (skill && skill.tools_enabled) {
        skill.tools_enabled[toolId] = newState
      }
      return newState
    } catch (e) {
      console.error('Failed to toggle skill:', e)
      throw e
    }
  }

  async function deleteSkill(skillId: string) {
    try {
      await App.DeleteSkill(skillId)
      // Remove from local state
      const index = skills.value.findIndex(s => s.id === skillId)
      if (index !== -1) {
        skills.value.splice(index, 1)
      }
    } catch (e) {
      console.error('Failed to delete skill:', e)
      throw e
    }
  }

  async function installFromPath(path: string, options: skill.InstallOptions) {
    try {
      const newSkill = await App.InstallFromPath(path, options)
      if (newSkill) {
        skills.value.push(newSkill)
      }
      return newSkill
    } catch (e) {
      console.error('Failed to install skill:', e)
      throw e
    }
  }

  async function updateSkill(skillId: string) {
    try {
      await App.UpdateSingleSkill(skillId)
      // Reload skills to get updated state
      await loadSkills()
    } catch (e) {
      console.error('Failed to update skill:', e)
      throw e
    }
  }

  function setCategory(category: string) {
    selectedCategory.value = category
  }

  function toggleTag(tag: string) {
    const index = selectedTags.value.indexOf(tag)
    if (index === -1) {
      selectedTags.value.push(tag)
    } else {
      selectedTags.value.splice(index, 1)
    }
  }

  function clearTagFilter() {
    selectedTags.value = []
  }

  function toggleToolFilter(toolId: string) {
    const index = selectedTools.value.indexOf(toolId)
    if (index === -1) {
      selectedTools.value.push(toolId)
    } else {
      selectedTools.value.splice(index, 1)
    }
  }

  function clearToolFilter() {
    selectedTools.value = []
  }

  // 分类管理方法
  async function loadCategories() {
    try {
      const result = await App.GetAllCategories()
      if (result) {
        customCategories.value = result
          .filter((c: CategoryInfo) => !c.is_preset)
          .map((c: CategoryInfo) => c.name)
      }
    } catch (e) {
      console.error('Failed to load categories:', e)
    }
  }

  async function addCategory(name: string) {
    try {
      await App.AddCategory(name)
      customCategories.value.push(name)
    } catch (e) {
      console.error('Failed to add category:', e)
      throw e
    }
  }

  async function deleteCategory(name: string): Promise<string[]> {
    try {
      const affectedSkills = await App.DeleteCategory(name)
      // 从本地状态移除
      const index = customCategories.value.indexOf(name)
      if (index !== -1) {
        customCategories.value.splice(index, 1)
      }
      // 如果删除的分类是当前选中的，重置为"全部"
      if (selectedCategory.value === name) {
        selectedCategory.value = '全部'
      }
      return affectedSkills || []
    } catch (e) {
      console.error('Failed to delete category:', e)
      throw e
    }
  }

  return {
    skills,
    tools,
    customTools,
    categories,
    customCategories,
    allCategoriesWithInfo,
    selectedCategory,
    selectedTags,
    searchQuery,
    isLoading,
    error,
    allTags,
    filteredSkills,
    totalSkills,
    enabledSkills,
    allEnabledTools,
    loadSkills,
    loadTools,
    loadCustomTools,
    toggleTool,
    deleteSkill,
    installFromPath,
    updateSkill,
    setCategory,
    toggleTag,
    clearTagFilter,
    selectedTools,
    toggleToolFilter,
    clearToolFilter,
    loadCategories,
    addCategory,
    deleteCategory
  }
})
