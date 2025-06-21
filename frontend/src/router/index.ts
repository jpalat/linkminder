import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: Dashboard,
    },
    {
      path: '/project/:id',
      name: 'project-detail',
      // route level code-splitting for project detail page
      component: () => import('../views/ProjectDetail.vue'),
    },
  ],
})

export default router
