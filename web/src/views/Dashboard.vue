<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { api, type Status } from '@/api'

const status = ref<Status | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

async function refresh() {
  try {
    status.value = await api.status()
    error.value = null
  } catch (e: any) {
    error.value = e.message ?? String(e)
  } finally {
    loading.value = false
  }
}

const activePeers = computed(() => {
  if (!status.value) return 0
  const now = Math.floor(Date.now() / 1000)
  return status.value.peers.filter((p) => now - p.latest_handshake < 180).length
})

onMounted(refresh)
</script>

<template>
  <el-row :gutter="16">
    <el-col :span="24">
      <el-alert v-if="error" type="error" :closable="false" show-icon>
        {{ error }}
      </el-alert>
    </el-col>
  </el-row>

  <el-skeleton v-if="loading" :rows="4" animated />

  <template v-else-if="status">
    <el-alert
      type="warning"
      :closable="false"
      show-icon
      title="Router port forwarding required"
      style="margin-bottom: 16px"
    >
      Forward UDP port <strong>{{ status.listen_port }}</strong> on your router
      to this device so clients can connect from outside the LAN.
    </el-alert>

    <el-row :gutter="16">
      <el-col :xs="24" :sm="8">
        <el-card>
          <div class="stat-label">Listen port (UDP)</div>
          <div class="stat-value">{{ status.listen_port }}</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="8">
        <el-card>
          <div class="stat-label">Endpoint</div>
          <div class="stat-value">{{ status.app_domain }}:{{ status.listen_port }}</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="8">
        <el-card>
          <div class="stat-label">Active peers</div>
          <div class="stat-value">{{ activePeers }} / {{ status.peers.length }}</div>
        </el-card>
      </el-col>
    </el-row>
  </template>
</template>

<style scoped>
.stat-label {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  margin-bottom: 8px;
}
.stat-value {
  font-size: 24px;
  font-weight: 600;
  overflow-wrap: anywhere;
}
</style>
