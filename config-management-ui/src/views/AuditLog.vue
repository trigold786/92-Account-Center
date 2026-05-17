<template>
  <div>
    <el-card style="margin-bottom: 16px">
      <el-row :gutter="16">
        <el-col :span="5">
          <el-select v-model="filter.operation_type" placeholder="操作类型" clearable @change="loadLogs" style="width: 100%">
            <el-option label="创建" value="CREATE" />
            <el-option label="更新" value="UPDATE" />
            <el-option label="删除" value="DELETE" />
            <el-option label="发布" value="RELEASE" />
          </el-select>
        </el-col>
        <el-col :span="5">
          <el-input v-model="filter.operator" placeholder="操作人" clearable @input="loadLogs" />
        </el-col>
        <el-col :span="5">
          <el-date-picker v-model="dateRange" type="datetimerange" range-separator="至" start-placeholder="开始时间" end-placeholder="结束时间" style="width: 100%" @change="onDateChange" />
        </el-col>
        <el-col :span="4" style="text-align: right">
          <el-button type="primary" @click="loadLogs">查询</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-card>
      <el-table :data="logs" stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="created_at" label="操作时间" width="180" />
        <el-table-column prop="operator" label="操作人" width="100" />
        <el-table-column prop="operator_ip" label="IP" width="130" />
        <el-table-column prop="operation_type" label="操作类型" width="120">
          <template #default="{ row }">
            <el-tag :type="typeTag(row.operation_type)" size="small">{{ row.operation_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="operation_object" label="操作对象" min-width="180" />
        <el-table-column prop="operation_result" label="结果" width="80">
          <template #default="{ row }">
            <el-tag :type="row.operation_result === 'success' ? 'success' : 'danger'" size="small">
              {{ row.operation_result }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button size="small" text @click="showDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="total > pageSize" style="text-align: center; margin-top: 16px">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          @current-change="loadLogs"
        />
      </div>
    </el-card>

    <el-dialog v-model="detailDialog" title="审计日志详情" width="600px">
      <template v-if="detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
          <el-descriptions-item label="操作时间">{{ detail.created_at }}</el-descriptions-item>
          <el-descriptions-item label="操作人">{{ detail.operator }}</el-descriptions-item>
          <el-descriptions-item label="IP地址">{{ detail.operator_ip }}</el-descriptions-item>
          <el-descriptions-item label="操作类型">{{ detail.operation_type }}</el-descriptions-item>
          <el-descriptions-item label="操作结果">{{ detail.operation_result }}</el-descriptions-item>
          <el-descriptions-item label="操作对象" :span="2">{{ detail.operation_object }}</el-descriptions-item>
          <el-descriptions-item label="操作详情" :span="2">{{ detail.operation_details }}</el-descriptions-item>
          <el-descriptions-item label="SM3哈希" :span="2">
            <code style="font-size: 11px">{{ detail.sm3_hash }}</code>
          </el-descriptions-item>
        </el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listAuditLogs, getAuditLog } from '@/api/audit'
import type { AuditLog } from '@/types'

const logs = ref<AuditLog[]>([])
const detail = ref<AuditLog | null>(null)
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const detailDialog = ref(false)
const dateRange = ref<[Date, Date] | null>(null)

const filter = ref({
  operation_type: '',
  operator: '',
  start_time: undefined as number | undefined,
  end_time: undefined as number | undefined,
})

function typeTag(t: string) {
  if (t.includes('CREATE')) return 'success'
  if (t.includes('UPDATE') || t.includes('EDIT')) return 'warning'
  if (t.includes('DELETE')) return 'danger'
  if (t.includes('RELEASE')) return 'primary'
  return 'info'
}

onMounted(() => loadLogs())

async function loadLogs() {
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (filter.value.operation_type) params.operation_type = filter.value.operation_type
    if (filter.value.operator) params.operator = filter.value.operator
    if (filter.value.start_time) params.start_time = filter.value.start_time
    if (filter.value.end_time) params.end_time = filter.value.end_time
    const res = await listAuditLogs(params)
    logs.value = res.data || []
    total.value = (res as any).total || 0
  } catch { /* ignore */ }
}

function onDateChange(val: [Date, Date] | null) {
  if (val) {
    filter.value.start_time = Math.floor(val[0].getTime() / 1000)
    filter.value.end_time = Math.floor(val[1].getTime() / 1000)
  } else {
    filter.value.start_time = undefined
    filter.value.end_time = undefined
  }
  loadLogs()
}

async function showDetail(row: AuditLog) {
  try {
    const res = await getAuditLog(row.id)
    detail.value = res.data || null
  } catch {
    detail.value = row
  }
  detailDialog.value = true
}
</script>
