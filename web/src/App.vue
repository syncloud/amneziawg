<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <el-container class="app-container">
    <el-header class="app-header" height="auto">
      <div class="app-title">AmneziaWG</div>
      <el-menu
        :default-active="route.name as string"
        mode="horizontal"
        router
        :ellipsis="false"
        class="app-menu"
      >
        <el-menu-item index="dashboard" :route="{ name: 'dashboard' }">Dashboard</el-menu-item>
        <el-menu-item index="peers" :route="{ name: 'peers' }">Peers</el-menu-item>
        <el-menu-item index="settings" :route="{ name: 'settings' }">Settings</el-menu-item>
        <el-menu-item index="logs" :route="{ name: 'logs' }">Logs</el-menu-item>
      </el-menu>
    </el-header>
    <el-main class="app-main">
      <router-view />
    </el-main>
  </el-container>
</template>

<style>
/* Global resets + mobile/desktop utility classes used across views. */
*,
*::before,
*::after {
  box-sizing: border-box;
}
html,
body {
  margin: 0;
  padding: 0;
  overflow-x: hidden;
}
.desktop-only {
  display: block;
}
.mobile-only {
  display: none;
}
@media (max-width: 768px) {
  .desktop-only {
    display: none;
  }
  .mobile-only {
    display: block;
  }
}
</style>

<style scoped>
.app-container {
  min-height: 100vh;
}
.app-header {
  display: flex;
  align-items: center;
  gap: 24px;
  flex-wrap: wrap;
  padding: 12px 20px;
  border-bottom: 1px solid var(--el-border-color-light);
}
.app-title {
  font-weight: 600;
  font-size: 18px;
  flex-shrink: 0;
}
.app-menu {
  flex: 1;
  border-bottom: none;
  min-width: 0;
}
.app-main {
  padding: 16px;
}
@media (min-width: 769px) {
  .app-main {
    padding: 24px;
  }
}
</style>
