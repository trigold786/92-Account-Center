import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: MainLayout,
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/Dashboard.vue'),
          meta: { title: '配置总览' },
        },
        {
          path: 'config',
          name: 'ConfigManagement',
          component: () => import('@/views/ConfigManagement.vue'),
          meta: { title: '配置管理' },
        },
        {
          path: 'config/edit/:id?',
          name: 'ConfigEdit',
          component: () => import('@/views/ConfigEdit.vue'),
          meta: { title: '编辑配置' },
        },
        {
          path: 'release',
          name: 'ReleaseApproval',
          component: () => import('@/views/ReleaseApproval.vue'),
          meta: { title: '发布审批' },
        },
        {
          path: 'audit',
          name: 'AuditLog',
          component: () => import('@/views/AuditLog.vue'),
          meta: { title: '审计日志' },
        },
        {
          path: 'permission',
          name: 'PermissionManage',
          component: () => import('@/views/PermissionManage.vue'),
          meta: { title: '权限管理' },
        },
      ],
    },
  ],
})

export default router
