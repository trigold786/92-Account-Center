<template>
  <div>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="概览" name="overview" v-if="hasPermission('data.dashboard')">
        <el-row :gutter="16">
          <el-col :xs="12" :sm="12" :md="6"><el-card><h3>注册用户</h3><p style="font-size:32px;color:#6C63FF">{{ overview?.total_users ?? '--' }}</p></el-card></el-col>
          <el-col :xs="12" :sm="12" :md="6"><el-card><h3>活跃用户</h3><p style="font-size:32px;color:#00D4FF">{{ overview?.active_users ?? '--' }}</p></el-card></el-col>
          <el-col :xs="12" :sm="12" :md="6"><el-card><h3>今日新增</h3><p style="font-size:32px;color:#2ED573">{{ overview?.new_today ?? '--' }}</p></el-card></el-col>
          <el-col :xs="12" :sm="12" :md="6"><el-card><h3>待发布</h3><p style="font-size:32px;color:#FFA502">{{ overview?.pending_releases ?? '--' }}</p></el-card></el-col>
        </el-row>
        <el-card style="margin-top:16px"><template #header>今日变更</template><p style="font-size:24px;color:#6C63FF">{{ overview?.today_changes ?? 0 }} 条审计</p></el-card>
      </el-tab-pane>

      <el-tab-pane label="用户管理" name="users" v-if="hasAnyRole(['admin', 'operator', 'support'])">
        <el-card>
          <template #header>
            <div style="display:flex;gap:12px;flex-wrap:wrap;align-items:center">
              <el-input v-model="userQuery" placeholder="手机号 / 账号" aria-label="搜索用户" style="width:240px" @keyup.enter="searchUsers" />
              <el-select v-model="statusFilter" placeholder="状态" clearable aria-label="状态筛选" style="width:130px" @change="searchUsers">
                <el-option label="全部状态" value="" />
                <el-option label="正常" value="active" />
                <el-option label="冻结" value="frozen" />
                <el-option label="封禁" value="banned" />
              </el-select>
              <el-select v-model="tierFilter" placeholder="等级" clearable aria-label="等级筛选" style="width:130px" @change="searchUsers">
                <el-option label="全部等级" :value="undefined" />
                <el-option label="普通" :value="0" />
                <el-option label="Silver" :value="1" />
                <el-option label="基础版" :value="2" />
                <el-option label="专业版" :value="3" />
                <el-option label="企业版" :value="4" />
              </el-select>
              <el-button @click="searchUsers" v-if="hasPermission('admin.user.manage')">查询</el-button>
            </div>
          </template>
          <div style="overflow-x: auto; width: 100%">
          <el-table :data="userList" stripe v-loading="userLoading">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="account_id" label="账号" width="140" />
            <el-table-column prop="phone_number" label="手机号" width="130" />
            <el-table-column prop="identity_tier" label="等级" width="80" />
            <el-table-column label="状态" width="90">
              <template #default="{ row }"><el-tag :type="row.status === 'active' ? 'success' : row.status === 'frozen' ? 'warning' : 'danger'">{{ row.status || 'active' }}</el-tag></template>
            </el-table-column>
            <el-table-column prop="created_at" label="注册时间" width="170" />
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button size="small" @click="toggleFreeze(row)" v-if="hasPermission('admin.user.freeze')">{{ row.status === 'frozen' ? '解冻' : '冻结' }}</el-button>
                <el-button size="small" type="danger" @click="toggleBan(row)" v-if="hasPermission('admin.user.ban')">{{ row.status === 'banned' ? '解封' : '封禁' }}</el-button>
                <el-button size="small" @click="adjustTier(row)" v-if="hasPermission('admin.user.manage')">等级</el-button>
              </template>
            </el-table-column>
          </el-table>
          </div>
        </el-card>
        <el-dialog v-model="showTierDialog" title="调整等级" :width="isMobile ? '95%' : '360px'">
          <el-select v-model="tierValue" placeholder="选择等级" style="width:100%">
            <el-option label="普通" :value="0" />
            <el-option label="Silver" :value="1" />
            <el-option label="基础版" :value="2" />
            <el-option label="专业版" :value="3" />
            <el-option label="企业版" :value="4" />
          </el-select>
          <el-input v-model="tierReason" placeholder="调整原因（必填）" style="width:100%;margin-top:12px" />
          <template #footer><el-button @click="showTierDialog = false">取消</el-button><el-button type="primary" @click="saveTier">确认</el-button></template>
        </el-dialog>
      </el-tab-pane>

      <el-tab-pane label="实名审核" name="kyc" v-if="hasAnyRole(['admin', 'operator'])">
        <el-card>
          <template #header>
            <div style="display:flex;justify-content:space-between;align-items:center">
              <span>待审核企业实名认证</span>
              <el-button size="small" @click="loadKYC" v-if="hasPermission('admin.user.manage')">刷新</el-button>
            </div>
          </template>
          <div style="overflow-x: auto; width: 100%">
          <el-table :data="kycList" stripe v-loading="kycLoading">
            <el-table-column prop="company_name" label="公司名称" />
            <el-table-column prop="unified_social_credit_code" label="统一社会信用代码" width="220" />
            <el-table-column prop="legal_person_name" label="法人" width="100" />
            <el-table-column prop="created_at" label="提交时间" width="180" />
            <el-table-column label="操作" width="180">
              <template #default="{ row }">
                <el-button size="small" type="success" @click="reviewEnterprise(row.enterprise_id, 'approve')" v-if="hasPermission('admin.user.manage')">通过</el-button>
                <el-button size="small" type="danger" @click="reviewEnterprise(row.enterprise_id, 'reject')" v-if="hasPermission('admin.user.manage')">拒绝</el-button>
              </template>
            </el-table-column>
          </el-table>
          </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="风险历史" name="risk" v-if="hasAnyRole(['admin', 'operator', 'support'])">
        <el-card><template #header>
          <div style="display:flex;gap:12px">
            <el-input v-model="riskUserId" placeholder="用户ID" style="width:200px" />
            <el-button @click="loadRisk" v-if="hasPermission('admin.risk.view')">查询</el-button>
          </div>
        </template>
          <div style="overflow-x: auto; width: 100%">
          <el-table :data="riskEvents" stripe>
            <el-table-column prop="event_type" label="事件类型" />
            <el-table-column prop="risk_level" label="风险等级" width="100">
              <template #default="{ row }"><el-tag :type="row.risk_level === 'high' ? 'danger' : row.risk_level === 'medium' ? 'warning' : 'success'">{{ row.risk_level }}</el-tag></template>
            </el-table-column>
            <el-table-column prop="risk_score" label="分值" width="80" />
            <el-table-column prop="created_at" label="时间" width="180" />
          </el-table>
          </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="黑名单" name="blacklist" v-if="hasAnyRole(['admin', 'operator'])">
        <el-card><template #header>
          <div style="display:flex;gap:12px">
            <el-select v-model="blType" placeholder="类型" style="width:120px">
              <el-option label="IP" value="ip" /><el-option label="设备" value="device" /><el-option label="用户" value="user" />
            </el-select>
            <el-input v-model="blValue" placeholder="值" style="width:200px" />
            <el-input v-model="blReason" placeholder="原因" style="width:200px" />
            <el-button type="primary" @click="addBlacklist" v-if="hasPermission('admin.blacklist.add')">添加</el-button>
          </div>
        </template>
          <div style="overflow-x: auto; width: 100%">
          <el-table :data="blacklist" stripe>
            <el-table-column prop="type" label="类型" width="100" />
            <el-table-column prop="value" label="值" />
            <el-table-column prop="reason" label="原因" />
            <el-table-column prop="created_at" label="时间" width="180" />
            <el-table-column label="操作" width="100">
              <template #default="{ row }"><el-button size="small" type="danger" @click="removeBl(row.type, row.value)" v-if="hasPermission('admin.blacklist.delete')">删除</el-button></template>
            </el-table-column>
          </el-table>
          </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="审计日志" name="audit" v-if="hasAnyRole(['admin', 'operator'])">
        <el-card>
        <div style="overflow-x: auto; width: 100%">
          <el-table :data="auditLogs" stripe>
            <el-table-column prop="event_type" label="事件" />
            <el-table-column prop="user_id" label="用户" width="100" />
            <el-table-column prop="details" label="详情" />
            <el-table-column prop="created_at" label="时间" width="180" />
            <el-table-column label="完整性" width="100"><template #default="{ row }"><el-button size="small" @click="verifyLog(row.id)" v-if="hasPermission('admin.audit.verify')">验证</el-button></template></el-table-column>
          </el-table>
        </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="SMS 提供商" name="sms" v-if="hasAnyRole(['admin', 'operator', 'support'])">
        <el-row :gutter="16" v-if="smsStatus">
          <el-col :span="8"><el-card><h3>提供商</h3><p style="font-size:20px;color:#00D4FF">{{ smsStatus.provider || smsStatus.name || '--' }}</p></el-card></el-col>
          <el-col :span="8"><el-card><h3>状态</h3><el-tag :type="(smsStatus.status === 'healthy' || smsStatus.healthy) ? 'success' : 'danger'" style="font-size:16px">{{ smsStatus.status || (smsStatus.healthy ? '正常' : '异常') }}</el-tag></el-card></el-col>
          <el-col :span="8"><el-card><h3>今日发送</h3><p style="font-size:20px;color:#2ED573">{{ smsStatus.sent_today ?? smsStatus.today_count ?? 0 }}</p></el-card></el-col>
          <el-col :span="12" style="margin-top:16px"><el-card><h3>配置</h3><pre style="color:var(--text-secondary);font-size:13px">{{ JSON.stringify(smsStatus.config || smsStatus, null, 2) }}</pre></el-card></el-col>
        </el-row>
        <el-card v-else><p style="color:var(--text-secondary);text-align:center;padding:20px">无法获取 SMS 提供商状态</p></el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { usePermissionStore } from '@/store/permission'
import { useBreakpoint } from '@/composables/useBreakpoint'
import { getRiskHistory, getAuditLogs, verifyAuditLog, listBlacklist, addBlacklistEntry, removeBlacklistEntry, getSMSProviderStatus, listUsers, updateUserStatus, updateUserTier, getPendingKYC, reviewKYC } from '@/api/admin'
import { getDashboardOverview } from '@/api/data'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { AdminUser, KYCStatus } from '@/types/api'

const perm = usePermissionStore()
const hasPermission = (p: string) => perm.hasPermission(p)
const hasAnyRole = (roles: string[]) => perm.hasAnyRole(roles)
const { isMobile } = useBreakpoint()

const activeTab = ref('overview')

watch(activeTab, (tab) => {
  if (tab === 'kyc') loadKYC()
})

// overview
interface Overview { total_users?: number; active_users?: number; new_today?: number; pending_releases?: number; today_changes?: number }
const overview = ref<Overview | null>(null)

// user management
const userQuery = ref('')
const userList = ref<AdminUser[]>([])
const userLoading = ref(false)
const statusFilter = ref('')
const tierFilter = ref<number | undefined>(undefined)
const showTierDialog = ref(false)
const tierUser = ref<AdminUser | null>(null)
const tierValue = ref<number>(0)
const tierReason = ref('')

// kyc
const kycList = ref<KYCStatus[]>([])
const kycLoading = ref(false)

// risk
interface RiskEvent { event_type: string; risk_level: string; risk_score: number; created_at: string }
const riskUserId = ref('')
const riskEvents = ref<RiskEvent[]>([])

// blacklist
interface BlacklistEntry { type: string; value: string; reason: string; created_at: string }
const blacklist = ref<BlacklistEntry[]>([])
const blType = ref('ip'); const blValue = ref(''); const blReason = ref('')

// audit
interface AuditLog { id: number; event_type: string; user_id: number; details: string; created_at: string }
const auditLogs = ref<AuditLog[]>([])

// sms
interface SMSStatus { provider?: string; name?: string; status?: string; healthy?: boolean; sent_today?: number; today_count?: number; config?: Record<string, unknown> }
const smsStatus = ref<SMSStatus | null>(null)

onMounted(async () => {
  try { const r = await getDashboardOverview(); overview.value = r.data.data } catch {}
  try { const r = await getSMSProviderStatus(); smsStatus.value = r.data.data } catch {}
  try {
    const now = new Date(); const weekAgo = new Date(now.getTime() - 7 * 86400000)
    const r = await getAuditLogs({ start_time: weekAgo.toISOString(), end_time: now.toISOString(), limit: 20 })
    auditLogs.value = r.data.data?.logs || r.data.data || []
  } catch {}
  try { const r = await listBlacklist({ limit: 50 }); blacklist.value = r.data.data || [] } catch {}
})

async function searchUsers() {
  userLoading.value = true
  try {
    const params: { search?: string; status?: string; tier?: number } = {}
    if (userQuery.value) params.search = userQuery.value
    if (statusFilter.value) params.status = statusFilter.value
    if (typeof tierFilter.value === 'number') params.tier = tierFilter.value
    const r = await listUsers(params)
    userList.value = r.data.data?.users || r.data.data || []
  } catch (e: any) { ElMessage.error(e.message) }
  userLoading.value = false
}

async function toggleFreeze(row: AdminUser) {
  const newStatus = row.status === 'frozen' ? 'active' : 'frozen'
  try { await updateUserStatus(row.id, newStatus, newStatus === 'frozen' ? '管理员冻结' : '管理员解冻'); ElMessage.success(newStatus === 'frozen' ? '已冻结' : '已解冻'); await searchUsers() } catch (e: any) { ElMessage.error(e.message) }
}

async function toggleBan(row: AdminUser) {
  const newStatus = row.status === 'banned' ? 'active' : 'banned'
  try { await updateUserStatus(row.id, newStatus, newStatus === 'banned' ? '管理员封禁' : '管理员解封'); ElMessage.success(newStatus === 'banned' ? '已封禁' : '已解封'); await searchUsers() } catch (e: any) { ElMessage.error(e.message) }
}

function adjustTier(row: AdminUser) {
  tierUser.value = row; tierValue.value = row.identity_tier ?? 0; tierReason.value = ''; showTierDialog.value = true
}

async function saveTier() {
  if (!tierUser.value) return
  if (!tierReason.value) { ElMessage.warning('请填写调整原因'); return }
  try { await updateUserTier(tierUser.value.id, tierValue.value, tierReason.value); ElMessage.success('等级已更新'); showTierDialog.value = false; await searchUsers() } catch (e: any) { ElMessage.error(e.message) }
}

async function loadKYC() {
  kycLoading.value = true
  try { const r = await getPendingKYC(); kycList.value = r.data.data?.entries || [] } catch (e: any) { ElMessage.error(e.message) }
  kycLoading.value = false
}

async function reviewEnterprise(id: string, action: 'approve' | 'reject') {
  try {
    await ElMessageBox.confirm(action === 'approve' ? '确认通过该企业实名认证？' : '确认拒绝该企业实名认证？', '提示', { type: action === 'approve' ? 'success' : 'warning' })
  } catch { return }
  try { await reviewKYC(id, action); ElMessage.success(action === 'approve' ? '已通过' : '已拒绝'); await loadKYC() } catch (e: any) { ElMessage.error(e.message) }
}

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
  try { const r = await verifyAuditLog(String(id)); ElMessage.success(r.data.data?.valid ? '日志完整' : '日志已被篡改') } catch (e: any) { ElMessage.error(e.message) }
}
</script>
