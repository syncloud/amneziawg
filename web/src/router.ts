import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'dashboard', component: () => import('./views/Dashboard.vue') },
    { path: '/peers', name: 'peers', component: () => import('./views/Peers.vue') },
    { path: '/settings', name: 'settings', component: () => import('./views/Settings.vue') },
    { path: '/logs', name: 'logs', component: () => import('./views/Logs.vue') },
  ],
})

export default router
