<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api, type Peer } from '@/api'

const peers = ref<Peer[]>([])
const loading = ref(true)
const addDialogOpen = ref(false)
const newPeerName = ref('')
const creating = ref(false)
const qrPeerId = ref<number | null>(null)

async function refresh() {
  loading.value = true
  try {
    peers.value = await api.listPeers()
  } catch (e: any) {
    ElMessage.error(e.message ?? String(e))
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
    ElMessage.success(`Created ${peer.name}`)
    // The immediate response carries the private key once; download now.
    downloadConfig(peer.id, peer.name)
    addDialogOpen.value = false
    newPeerName.value = ''
    await refresh()
  } catch (e: any) {
    ElMessage.error(e.message ?? String(e))
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
    ElMessage.success(`Removed ${peer.name}`)
    await refresh()
  } catch (e: any) {
    ElMessage.error(e.message ?? String(e))
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

onMounted(refresh)
</script>

<template>
  <el-row justify="space-between" align="middle" style="margin-bottom: 16px">
    <h2 style="margin: 0">Peers</h2>
    <el-button type="primary" @click="addDialogOpen = true">Add peer</el-button>
  </el-row>

  <el-table :data="peers" v-loading="loading" empty-text="No peers yet">
    <el-table-column prop="name" label="Name" />
    <el-table-column prop="address_v4" label="Address" width="180" />
    <el-table-column prop="created_at" label="Created" width="200" />
    <el-table-column label="Actions" width="320">
      <template #default="{ row }">
        <el-button size="small" @click="downloadConfig(row.id, row.name)">Config</el-button>
        <el-button size="small" @click="qrPeerId = row.id">QR</el-button>
        <el-button size="small" type="danger" @click="remove(row)">Remove</el-button>
      </template>
    </el-table-column>
  </el-table>

  <el-dialog v-model="addDialogOpen" title="Add peer" width="400px">
    <el-form @submit.prevent="create">
      <el-form-item label="Name">
        <el-input v-model="newPeerName" placeholder="laptop, phone, …" autofocus />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="addDialogOpen = false">Cancel</el-button>
      <el-button type="primary" :loading="creating" @click="create">Create</el-button>
    </template>
  </el-dialog>

  <el-dialog v-model="qrPeerId" :title="`QR — peer #${qrPeerId}`" width="540px">
    <img v-if="qrPeerId" :src="api.peerQRURL(qrPeerId)" alt="QR code" style="width: 100%" />
  </el-dialog>
</template>
