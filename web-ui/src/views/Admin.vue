<template>
  <div>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="概览" name="overview">
        <el-card><template #header>数据概览</template>
          <pre style="color:var(--text-secondary)">{{ JSON.stringify(overview, null, 2) }}</pre>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="风险历史" name="risk">
        <el-card><template #header>
          <div style="display:flex;gap:12px">
            <el-input v-model="riskUserId" placeholder="用户ID" style="width:200px" />
            <el-button @click="loadRisk">查询</el-button>
          </div>
        </template>
          <el-table :data="riskEvents" stripe>
            <el-table-column prop="event_type" label="事件类型" />
            <el-table-column prop="risk_level" label="风险等级" width="100">
              <template #default="{ row }"><el-tag :type="row.risk_level === 'high' ? 'danger' : row.risk_level === 'medium' ? 'warning' : 'success'">{{ row.risk_level }}</el-tag></template>
            </el-table-column>
            <el-table-column prop="risk_score" label="分值" width="80" />
            <el-table-column prop="created_at" label="时间" width="180" />
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="黑名单" name="blacklist">
        <el-card><template #header>
          <div style="display:flex;gap:12px">
            <el-select v-model="blType" placeholder="类型" style="width:120px">
              <el-option label="IP" value="ip" /><el-option label="设备" value="device" /><el-option label="用户" value="user" />
            </el-select>
            <el-input v-model="blValue" placeholder="值" style="width:200px" />
            <el-input v-model="blReason" placeholder="原因" style="width:200px" />
            <el-button type="primary" @click="addBlacklist">添加</el-button>
          </div>
        </template>
          <el-table :data="blacklist" stripe>
            <el-table-column prop="type" label="类型" width="100" />
            <el-table-column prop="value" label="值" />
            <el-table-column prop="reason" label="原因" />
            <el-table-column prop="created_at" label="时间" width="180" />
            <el-table-column label="操作" width="100">
              <template #default="{ row }"><el-button size="small" type="danger" @click="removeBl(row.type, row.value)">删除</el-button></template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="审计日志" name="audit">
        <el-card>
          <el-table :data="auditLogs" stripe>
            <el-table-column prop="event_type" label="事件" />
            <el-table-column prop="user_id" label="用户" width="100" />
            <el-table-column prop="details" label="详情" />
            <el-table-column prop="created_at" label="时间" width="180" />
            <el-table-column label="完整性" width="100"><template #default="{ row }"><el-button size="small" @click="verifyLog(row.id)">验证</el-button></template></el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="SMS 提供商" name="sms">
        <el-card><pre style="color:var(--text-secondary)">{{ JSON.stringify(smsStatus, null, 2) }}</pre></el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getRiskHistory, getAuditLogs, verifyAuditLog, listBlacklist, addBlacklistEntry, removeBlacklistEntry, getSMSProviderStatus } from '@/api/admin'
import { getDashboardOverview } from '@/api/data'
import { ElMessage } from 'element-plus'

const activeTab = ref('overview')
const overview = ref<any>(null)
const riskUserId = ref('')
const riskEvents = ref<any[]>([])
const blacklist = ref<any[]>([])
const auditLogs = ref<any[]>([])
const smsStatus = ref<any>(null)
const blType = ref('ip'); const blValue = ref(''); const blReason = ref('')

onMounted(async () => {
  try { const r = await getDashboardOverview(); overview.value = r.data.data } catch {}
  try { const r = await getSMSProviderStatus(); smsStatus.value = r.data.data } catch {}
  try {
    const now = new Date()
    const weekAgo = new Date(now.getTime() - 7 * 86400000)
    const r = await getAuditLogs({ start_time: weekAgo.toISOString(), end_time: now.toISOString(), limit: 20 })
    auditLogs.value = r.data.data?.logs || r.data.data || []
  } catch {}
  try { const r = await listBlacklist({ limit: 50 }); blacklist.value = r.data.data || [] } catch {}
})

async function loadRisk() {
  if (!riskUserId.value) return
  try { const r = await getRiskHistory(Number(riskUserId.value)); riskEvents.value = r.data.data || [] } catch (e: any) { ElMessage.error(e.message) }
}

async function addBlacklist() {
  if (!blValue.value) return
  try { await addBlacklistEntry({ type: blType.value, value: blValue.value, reason: blReason.value }); ElMessage.success('已添加'); blValue.value = ''; blReason.value = ''; const r = await listBlacklist({ page_size: 50 }); blacklist.value = r.data.data || [] } catch (e: any) { ElMessage.error(e.message) }
}

async function removeBl(type: string, value: string) {
  try { await removeBlacklistEntry(type, value); ElMessage.success('已删除'); const r = await listBlacklist({ page_size: 50 }); blacklist.value = r.data.data || [] } catch (e: any) { ElMessage.error(e.message) }
}

async function verifyLog(id: number) {
  try { const r = await verifyAuditLog(id); ElMessage.success(r.data.data?.valid ? '日志完整' : '日志已被篡改') } catch (e: any) { ElMessage.error(e.message) }
}
</script>
