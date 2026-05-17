<template>
  <div>
    <el-row :gutter="20" style="margin-bottom: 20px">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value">{{ stats.total_config }}</div>
            <div class="stat-label">总配置项</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value" style="color: #e6a23c">{{ stats.pending_releases }}</div>
            <div class="stat-label">待审批发布单</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value" style="color: #409eff">{{ stats.today_changes }}</div>
            <div class="stat-label">今日变更</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-value" :style="{ color: stats.alert_count > 0 ? '#f56c6c' : '#67c23a' }">
              {{ stats.alert_count }}
            </div>
            <div class="stat-label">异常告警</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>最近变更记录</span>
          <el-button text type="primary" @click="$router.push('/audit')">查看全部</el-button>
        </div>
      </template>
      <el-table :data="recentChanges" stripe style="width: 100%">
        <el-table-column prop="created_at" label="时间" width="180" />
        <el-table-column prop="operator" label="操作人" width="120" />
        <el-table-column prop="operation_type" label="操作类型" width="120">
          <template #default="{ row }">
            <el-tag :type="typeTag(row.operation_type)" size="small">{{ row.operation_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="operation_object" label="操作对象" />
        <el-table-column prop="operation_result" label="结果" width="80">
          <template #default="{ row }">
            <el-tag :type="row.operation_result === 'success' ? 'success' : 'danger'" size="small">
              {{ row.operation_result }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listAuditLogs } from '@/api/audit'
import { getStats } from '@/api/config'

const stats = ref({
  total_config: 0,
  enabled_config: 0,
  pending_releases: 0,
  today_changes: 0,
  alert_count: 0,
})

const recentChanges = ref<any[]>([])

function typeTag(type: string) {
  if (type.includes('CREATE')) return 'success'
  if (type.includes('UPDATE') || type.includes('EDIT')) return 'warning'
  if (type.includes('DELETE')) return 'danger'
  return 'info'
}

onMounted(async () => {
  try {
    const statsRes = await getStats()
    if (statsRes.code === 0) {
      stats.value = { ...stats.value, ...statsRes.data }
    }
  } catch (e) {
    console.warn('Failed to load dashboard stats:', e)
  }

  try {
    const res = await listAuditLogs({ page: 1, page_size: 10 })
    recentChanges.value = (res as any).data || []
  } catch (e) {
    console.warn('Failed to load recent audit logs:', e)
  }
})
</script>

<style scoped>
.stat-card {
  text-align: center;
  padding: 10px;
}
.stat-value {
  font-size: 36px;
  font-weight: bold;
}
.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 8px;
}
</style>
