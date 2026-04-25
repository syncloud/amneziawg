<script setup lang="ts">
import { notify } from '@/notify'

const isStub = !!import.meta.env.VITE_STUB

function stubLogout(e: MouseEvent) {
  if (!isStub) return
  e.preventDefault()
  notify.info('Logout stubbed in dev — wired to /auth/logout in production')
}
</script>

<template>
  <el-container class="app-container">
    <el-header class="app-header" height="auto">
      <div class="header-top">
        <div class="app-title">
          <span class="brand-dot" />
          <span class="brand-text">AmneziaWG</span>
        </div>
        <a href="/auth/logout" class="logout-link" data-testid="logout" @click="stubLogout">
          <el-button text>Logout</el-button>
        </a>
      </div>
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
  display: flex;
  align-items: center;
  gap: 10px;
  font-weight: 600;
  font-size: 18px;
  letter-spacing: 0.2px;
}
.brand-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--awg-connected);
  box-shadow: 0 0 0 4px rgba(76, 175, 80, 0.16), 0 0 10px rgba(76, 175, 80, 0.6);
}
.brand-text {
  color: var(--awg-text);
}
.logout-link {
  text-decoration: none;
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
    padding: 0 12px;
  }
  .header-top {
    padding: 10px 0;
  }
}
</style>
