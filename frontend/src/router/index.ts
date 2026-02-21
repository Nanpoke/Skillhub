import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../views/HomeView.vue')
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('../views/SettingsView.vue')
    },
    {
      path: '/install',
      name: 'install',
      component: () => import('../views/InstallView.vue')
    },
    {
      path: '/install/git',
      name: 'install-git',
      component: () => import('../views/GitInstallView.vue')
    },
    {
      path: '/install/local',
      name: 'install-local',
      component: () => import('../views/LocalInstallView.vue')
    },
    {
      path: '/wizard',
      name: 'wizard',
      component: () => import('../views/WizardView.vue')
    },
    {
      path: '/viewer/:id',
      name: 'viewer',
      component: () => import('../views/ViewerView.vue')
    },
    {
      path: '/history',
      name: 'history',
      component: () => import('../views/HistoryView.vue')
    },
    {
      path: '/import-export',
      name: 'import-export',
      component: () => import('../views/ImportExportView.vue')
    }
  ]
})

export default router
