<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessageBox } from 'element-plus'
import QRCode from 'qrcode'
import { api, type Peer, type Status } from '@/api'
import { notify } from '@/notify'

const peers = ref<Peer[]>([])
const status = ref<Status | null>(null)
const loading = ref(true)
const filter = ref('')

const addDialogOpen = ref(false)
const newPeerName = ref('')
const creating = ref(false)

const qrPeerId = ref<number | null>(null)
const qrSrc = ref<string | null>(null)
const qrLoading = ref(false)
const qrPeerName = computed(() => peers.value.find((p) => p.id === qrPeerId.value)?.name ?? '')

const dismissedPort = ref<string | null>(localStorage.getItem('port-forward-dismissed'))

const ONLINE_THRESHOLD_S = 180

function stripPrefix(addr: string): string {
  const i = addr.indexOf('/')
  return i < 0 ? addr : addr.slice(0, i)
}

function isOnline(peer: Peer): boolean {
  const s = status.value?.peers.find((sp) => sp.public_key === peer.public_key)
  if (!s || !s.latest_handshake) return false
  return Math.floor(Date.now() / 1000) - s.latest_handshake < ONLINE_THRESHOLD_S
}

const filteredPeers = computed(() => {
  const f = filter.value.trim().toLowerCase()
  if (!f) return peers.value
  return peers.value.filter((p) => p.name.toLowerCase().includes(f))
})

const activeCount = computed(() => peers.value.filter(isOnline).length)

const showReminder = computed(() => {
  if (!status.value) return false
  return dismissedPort.value !== String(status.value.listen_port)
})

function dismissReminder() {
  if (!status.value) return
  const key = String(status.value.listen_port)
  localStorage.setItem('port-forward-dismissed', key)
  dismissedPort.value = key
}

async function refresh() {
  loading.value = true
  try {
    const [peerList, st] = await Promise.all([api.listPeers(), api.status()])
    peers.value = peerList
    status.value = st
  } catch (e: any) {
    notify.error(e.message ?? String(e))
  } finally {
    loading.value = false
  }
}

async function create() {
  const name = newPeerName.value.trim()
  if (!name) return
  creating.value = true
  try {
    const peer = await api.createPeer(name)
    notify.success(`Created ${peer.name}`)
    downloadConfig(peer.id, peer.name)
    addDialogOpen.value = false
    newPeerName.value = ''
    await refresh()
  } catch (e: any) {
    notify.error(e.message ?? String(e))
  } finally {
    creating.value = false
  }
}

async function remove(peer: Peer) {
  await ElMessageBox.confirm(
    `Remove peer "${peer.name}"? The client's config will stop working immediately.`,
    'Confirm remove',
    { type: 'warning' }
  )
  try {
    await api.deletePeer(peer.id)
    notify.success(`Removed ${peer.name}`)
    await refresh()
  } catch (e: any) {
    notify.error(e.message ?? String(e))
  }
}

function downloadConfig(id: number, name: string) {
  const a = document.createElement('a')
  a.href = api.peerConfigURL(id)
  a.download = `${name}.conf`
  document.body.appendChild(a)
  a.click()
  a.remove()
}

watch(qrPeerId, async (id) => {
  qrSrc.value = null
  if (id == null) return
  qrLoading.value = true
  try {
    const conf = await api.peerConfigText(id)
    qrSrc.value = await QRCode.toDataURL(conf, { width: 512, margin: 1 })
  } catch (e: any) {
    notify.error(e.message ?? String(e))
    qrPeerId.value = null
  } finally {
    qrLoading.value = false
  }
})

onMounted(refresh)
</script>

<template>
  <el-alert
    v-if="showReminder"
    type="warning"
    closable
    show-icon
    title="Router port forwarding required"
    style="margin-bottom: 16px"
    data-testid="port-forward-reminder"
    @close="dismissReminder"
  >
    Forward UDP port <strong data-testid="listen-port">{{ status?.listen_port }}</strong> on your router
    to this device so clients can connect from outside the LAN.
  </el-alert>

  <div class="summary" v-if="status">
    <div class="summary-item">
      <div class="summary-label">Endpoint</div>
      <div class="summary-value" data-testid="endpoint">{{ status.app_domain }}:{{ status.listen_port }}</div>
    </div>
    <div class="summary-item">
      <div class="summary-label">Active</div>
      <div class="summary-value">{{ activeCount }} / {{ peers.length }}</div>
    </div>
  </div>

  <div class="peers-toolbar">
    <el-input
      v-model="filter"
      placeholder="Filter by name"
      clearable
      data-testid="peer-filter"
      class="peer-filter"
    />
    <el-button type="primary" @click="addDialogOpen = true" data-testid="add-peer-button">Add peer</el-button>
  </div>

  <div class="desktop-only">
    <el-table :data="filteredPeers" v-loading="loading" empty-text="No peers">
      <el-table-column label="" width="48">
        <template #default="{ row }">
          <span class="online-dot" :class="{ on: isOnline(row) }" :title="isOnline(row) ? 'online' : 'offline'" />
        </template>
      </el-table-column>
      <el-table-column prop="name" label="Name" />
      <el-table-column label="Address" width="180">
        <template #default="{ row }">{{ stripPrefix(row.address_v4) }}</template>
      </el-table-column>
      <el-table-column prop="created_at" label="Created" width="200" />
      <el-table-column label="Actions" width="320">
        <template #default="{ row }">
          <el-button size="small" @click="downloadConfig(row.id, row.name)">Config</el-button>
          <el-button size="small" @click="qrPeerId = row.id">QR</el-button>
          <el-button size="small" type="danger" @click="remove(row)">Remove</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>

  <div class="mobile-only" v-loading="loading">
    <el-empty v-if="!loading && filteredPeers.length === 0" :description="peers.length === 0 ? 'No peers' : 'No matches'" />
    <el-card v-for="peer in filteredPeers" :key="peer.id" class="peer-card" shadow="never">
      <div class="peer-card-header">
        <div class="peer-name">
          <span class="online-dot" :class="{ on: isOnline(peer) }" />
          {{ peer.name }}
        </div>
        <div class="peer-addr">{{ stripPrefix(peer.address_v4) }}</div>
      </div>
      <div class="peer-meta">Created {{ peer.created_at }}</div>
      <div class="peer-actions">
        <el-button size="small" @click="downloadConfig(peer.id, peer.name)">Config</el-button>
        <el-button size="small" @click="qrPeerId = peer.id">QR</el-button>
        <el-button size="small" type="danger" @click="remove(peer)">Remove</el-button>
      </div>
    </el-card>
  </div>

  <el-dialog v-model="addDialogOpen" title="Add peer" width="min(400px, 92vw)" data-testid="add-peer-dialog">
    <el-form @submit.prevent="create">
      <el-form-item label="Name">
        <el-input v-model="newPeerName" placeholder="laptop, phone, …" autofocus data-testid="peer-name-input" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="addDialogOpen = false" data-testid="add-peer-cancel">Cancel</el-button>
      <el-button type="primary" :loading="creating" @click="create" data-testid="add-peer-create">Create</el-button>
    </template>
  </el-dialog>

  <el-dialog
    :model-value="qrPeerId !== null"
    @update:model-value="(v: boolean) => { if (!v) qrPeerId = null }"
    :title="qrPeerName ? `QR — ${qrPeerName}` : 'QR'"
    width="min(540px, 92vw)"
  >
    <el-skeleton v-if="qrLoading" :rows="6" animated />
    <img v-else-if="qrSrc" :src="qrSrc" alt="QR code" style="width: 100%" data-testid="peer-qr" />
  </el-dialog>
</template>

<style scoped>
.summary {
  display: flex;
  gap: 24px;
  padding: 16px 20px;
  margin-bottom: 20px;
  background: var(--awg-surface);
  border: 1px solid var(--awg-border);
  border-radius: 12px;
  flex-wrap: wrap;
}
.summary-item {
  flex: 1 1 auto;
  min-width: 140px;
}
.summary-label {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  margin-bottom: 4px;
}
.summary-value {
  font-size: 18px;
  font-weight: 600;
  overflow-wrap: anywhere;
}
.peers-toolbar {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.peer-filter {
  flex: 1 1 220px;
  max-width: 360px;
}
.online-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--awg-text-muted);
  margin-right: 6px;
  vertical-align: middle;
}
.online-dot.on {
  background: var(--awg-connected);
  box-shadow: 0 0 0 3px rgba(76, 175, 80, 0.18), 0 0 8px rgba(76, 175, 80, 0.6);
}
.peer-card {
  margin-bottom: 12px;
}
.peer-card-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 6px;
}
.peer-name {
  font-weight: 600;
  font-size: 16px;
}
.peer-addr {
  font-family: var(--el-font-family-mono, monospace);
  font-size: 13px;
  color: var(--el-text-color-regular);
}
.peer-meta {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 12px;
}
.peer-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.peer-actions .el-button {
  flex: 1 1 auto;
}
</style>
