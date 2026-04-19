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
      </el-menu>
      <a href="/auth/logout" class="logout-link">
        <el-button text>Logout</el-button>
      </a>
    </el-header>
    <el-main class="app-main">
      <router-view />
    </el-main>
  </el-container>
</template>

<style>
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
.logout-link {
  text-decoration: none;
  flex-shrink: 0;
}
.app-main {
  padding: 16px;
}
@media (min-width: 769px) {
  .app-main {
    padding: 24px;
  }
}

@media (max-width: 768px) {
  .app-header {
    flex-direction: row;
    flex-wrap: wrap;
    align-items: center;
    gap: 0;
    padding: 8px 12px;
  }
  .app-title {
    padding: 4px 4px 8px;
    flex: 1;
  }
  .logout-link {
    margin-left: auto;
  }
  .app-menu {
    width: 100%;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    order: 3;
  }
  .app-menu :deep(.el-menu) {
    min-width: max-content;
  }
  .app-menu :deep(.el-menu-item) {
    padding: 0 14px;
    height: 44px;
    line-height: 44px;
    font-size: 14px;
  }
}
</style>
