import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../store/user'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('../views/Layout.vue'),
    meta: { requiresAuth: true },
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue'),
        meta: { title: '仪表盘', icon: 'DataAnalysis' }
      },
      {
        path: 'hosts',
        name: 'Hosts',
        component: () => import('../views/Hosts.vue'),
        meta: { title: '主机管理', icon: 'Monitor' }
      },
      {
        path: 'tasks',
        name: 'Tasks',
        component: () => import('../views/Tasks.vue'),
        meta: { title: '任务管理', icon: 'Calendar' }
      },
      {
        path: 'backups',
        name: 'Backups',
        component: () => import('../views/Backups.vue'),
        meta: { title: '备份管理', icon: 'Files' }
      },
      {
        path: 'storages',
        name: 'Storages',
        component: () => import('../views/Storages.vue'),
        meta: { title: '存储管理', icon: 'FolderOpened' }
      },
      {
        path: 'notifications',
        name: 'Notifications',
        component: () => import('../views/Notifications.vue'),
        meta: { title: '通知管理', icon: 'Bell' }
      },
      {
        path: 'logs',
        name: 'Logs',
        component: () => import('../views/Logs.vue'),
        meta: { title: '备份日志', icon: 'Document' }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('../views/Users.vue'),
        meta: { title: '用户管理', icon: 'User' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const userStore = useUserStore()

  if (to.meta.requiresAuth !== false && !userStore.isLoggedIn) {
    next('/login')
  } else if (to.path === '/login' && userStore.isLoggedIn) {
    next('/')
  } else {
    next()
  }
})

export default router
