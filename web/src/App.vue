<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<template>
  <el-container class="app-container">
    <el-header class="app-header">
      <div class="app-title">AmneziaWG</div>
      <el-menu
        :default-active="route.name as string"
        mode="horizontal"
        router
      >
        <el-menu-item index="dashboard" :route="{ name: 'dashboard' }">Dashboard</el-menu-item>
        <el-menu-item index="peers" :route="{ name: 'peers' }">Peers</el-menu-item>
        <el-menu-item index="settings" :route="{ name: 'settings' }">Settings</el-menu-item>
        <el-menu-item index="logs" :route="{ name: 'logs' }">Logs</el-menu-item>
      </el-menu>
    </el-header>
    <el-main>
      <router-view />
    </el-main>
  </el-container>
</template>

<style>
.app-container {
  min-height: 100vh;
}
.app-header {
  display: flex;
  align-items: center;
  gap: 32px;
  border-bottom: 1px solid var(--el-border-color-light);
}
.app-title {
  font-weight: 600;
  font-size: 18px;
}
</style>
