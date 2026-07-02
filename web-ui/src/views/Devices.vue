<template>
  <el-card>
    <template #header>登录设备</template>
    <div style="overflow-x: auto; width: 100%">
    <el-table :data="devices" stripe>
      <el-table-column prop="device_name" label="设备名称" />
      <el-table-column prop="device_type" label="类型" width="120" />
      <el-table-column prop="last_seen_at" label="最后活跃" width="180" />
      <el-table-column label="可信" width="80"><template #default="{ row }"><el-tag :type="row.is_trusted ? 'success' : 'info'">{{ row.is_trusted ? '是' : '否' }}</el-tag></template></el-table-column>
      <el-table-column label="操作" width="120">
        <template #default="{ row }">
          <el-button v-if="!row.is_trusted && hasPermission('device.trust')" size="small" @click="trustDevice(row.device_id)">信任</el-button>
          <el-button size="small" type="danger" @click="removeDevice(row.device_id)" v-if="hasPermission('device.remove')">移除</el-button>
        </template>
      </el-table-column>
    </el-table>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/store/auth'
import { usePermissionStore } from '@/store/permission'
import { getUserDevices, trustDevice as apiTrustDevice, removeDevice as apiRemoveDevice } from '@/api/device'
import { ElMessage } from 'element-plus'
import type { Device } from '@/types/api'

const auth = useAuthStore()
const perm = usePermissionStore()
const hasPermission = (p: string) => perm.hasPermission(p)
const devices = ref<Device[]>([])

onMounted(async () => {
  const uid = auth.userId; if (!uid) return
  try { const r = await getUserDevices(uid); devices.value = r.data.data || [] } catch {}
})

async function trustDevice(id: string) { try { await apiTrustDevice(id); ElMessage.success('已信任'); devices.value = devices.value.map(d => d.device_id === id ? { ...d, is_trusted: true } : d) } catch (e: any) { ElMessage.error(e.message) } }
async function removeDevice(id: string) { try { await apiRemoveDevice(id); ElMessage.success('已移除'); devices.value = devices.value.filter(d => d.device_id !== id) } catch (e: any) { ElMessage.error(e.message) } }
</script>
