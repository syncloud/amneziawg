<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <el-container class="app-container">
    <el-header class="app-header" height="auto">
      <div class="header-top">
        <div class="app-title">AmneziaWG</div>
        <a href="/auth/logout" class="logout-link">
          <el-button text>Logout</el-button>
        </a>
      </div>
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
  padding: 0 20px;
  border-bottom: 1px solid var(--el-border-color-light);
}
.header-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 0;
}
.app-title {
  font-weight: 600;
  font-size: 18px;
}
.logout-link {
  text-decoration: none;
}
.app-menu {
  border-bottom: none;
}
.app-main {
  padding: 16px;
}
@media (min-width: 769px) {
  .app-header {
    display: flex;
    align-items: center;
    gap: 24px;
  }
  .header-top {
    padding: 12px 0;
    flex: 0 0 auto;
  }
  .app-menu {
    flex: 1;
  }
  .app-main {
    padding: 24px;
  }
}
@media (max-width: 768px) {
  .app-header {
    padding: 0 12px;
  }
  .header-top {
    padding: 10px 0;
  }
  .app-menu :deep(.el-menu-item) {
    padding: 0 16px;
    height: 44px;
    line-height: 44px;
    font-size: 14px;
    flex: 1;
    text-align: center;
    justify-content: center;
  }
}
</style>
