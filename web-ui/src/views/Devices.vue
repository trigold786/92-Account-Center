<template>
  <el-card>
    <template #header>登录设备</template>
    <el-table :data="devices" stripe>
      <el-table-column prop="device_name" label="设备名称" />
      <el-table-column prop="device_type" label="类型" width="120" />
      <el-table-column prop="last_seen_at" label="最后活跃" width="180" />
      <el-table-column label="可信" width="80"><template #default="{ row }"><el-tag :type="row.is_trusted ? 'success' : 'info'">{{ row.is_trusted ? '是' : '否' }}</el-tag></template></el-table-column>
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <el-button v-if="!row.is_trusted" size="small" @click="trustDevice(row.device_id)">信任</el-button>
          <el-button size="small" type="danger" @click="removeDevice(row.device_id)">移除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/store/auth'
import { getUserDevices, trustDevice, removeDevice } from '@/api/device'
import { ElMessage } from 'element-plus'

const auth = useAuthStore()
const devices = ref<any[]>([])

onMounted(async () => {
  const uid = auth.userId; if (!uid) return
  try { const r = await getUserDevices(uid); devices.value = r.data.data || [] } catch {}
})

async function trustDevice(id: string) { try { await trustDevice(id); ElMessage.success('已信任'); devices.value = devices.value.map(d => d.device_id === id ? { ...d, is_trusted: true } : d) } catch (e: any) { ElMessage.error(e.message) } }
async function removeDevice(id: string) { try { await removeDevice(id); ElMessage.success('已移除'); devices.value = devices.value.filter(d => d.device_id !== id) } catch (e: any) { ElMessage.error(e.message) } }
</script>
