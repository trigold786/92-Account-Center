import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'Login', component: () => import('@/views/Login.vue'), meta: { layout: 'auth' } },
    { path: '/register', name: 'Register', component: () => import('@/views/Register.vue'), meta: { layout: 'auth' } },
    { path: '/', name: 'Dashboard', component: () => import('@/views/Dashboard.vue'), meta: { requiresAuth: true } },
    { path: '/account', name: 'Account', component: () => import('@/views/Account.vue'), meta: { requiresAuth: true } },
    { path: '/credits', name: 'Credits', component: () => import('@/views/Credits.vue'), meta: { requiresAuth: true } },
    { path: '/subscriptions', name: 'Subscriptions', component: () => import('@/views/Subscriptions.vue'), meta: { requiresAuth: true } },
    { path: '/referral', name: 'Referral', component: () => import('@/views/Referral.vue'), meta: { requiresAuth: true } },
    { path: '/devices', name: 'Devices', component: () => import('@/views/Devices.vue'), meta: { requiresAuth: true } },
    { path: '/admin', name: 'Admin', component: () => import('@/views/Admin.vue'), meta: { requiresAuth: true } },
    { path: '/:pathMatch(.*)*', name: 'NotFound', component: () => import('@/views/NotFound.vue') },
  ],
})

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('access_token')
  if (to.meta.requiresAuth && !token) next('/login')
  else if (to.path === '/login' && token) next('/')
  else next()
})

export default router
