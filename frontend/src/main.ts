import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './style.css'

// 等待 Wails 运行时就绪后再挂载应用
function initApp() {
  const app = createApp(App)

  app.use(createPinia())
  app.use(router)

  app.mount('#app')
}

// 检查 Wails 运行时是否就绪
if (window.runtime) {
  // 运行时已就绪，直接初始化
  initApp()
} else {
  // 等待运行时就绪
  // Wails 会在 window 对象上注入 runtime
  const checkRuntime = setInterval(() => {
    if (window.runtime) {
      clearInterval(checkRuntime)
      initApp()
    }
  }, 10)

  // 超时保护：最多等待 5 秒
  setTimeout(() => {
    clearInterval(checkRuntime)
    if (!window.runtime) {
      console.warn('Wails runtime not detected, initializing anyway...')
      initApp()
    }
  }, 5000)
}
