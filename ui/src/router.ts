import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { layout: 'blank', public: true },
  },
  {
    path: '/setup',
    name: 'setup',
    component: () => import('@/views/SetupView.vue'),
    meta: { layout: 'blank', public: true },
  },
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: () => import('@/views/DashboardView.vue'),
  },
  {
    path: '/items',
    name: 'items',
    component: () => import('@/views/ItemsView.vue'),
  },
  {
    path: '/items/:id',
    name: 'item-detail',
    component: () => import('@/views/ItemDetailView.vue'),
    props: true,
  },
  {
    path: '/sources',
    name: 'sources',
    component: () => import('@/views/SourcesView.vue'),
  },
  {
    path: '/rules',
    name: 'rules',
    component: () => import('@/views/RulesView.vue'),
  },
  {
    path: '/telegram',
    name: 'telegram',
    component: () => import('@/views/TelegramView.vue'),
  },
  {
    path: '/account',
    name: 'account',
    component: () => import('@/views/AccountView.vue'),
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/views/NotFoundView.vue'),
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  },
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.public) return true
  if (!auth.isAuthenticated) {
    return { name: 'login', query: { next: to.fullPath } }
  }
  return true
})

export default router
