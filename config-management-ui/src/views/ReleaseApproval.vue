<template>
  <div>
    <el-card style="margin-bottom: 16px">
      <el-row :gutter="16">
        <el-col :span="4">
          <el-select v-model="statusFilter" placeholder="按状态筛选" clearable @change="loadReleases" style="width: 100%">
            <el-option label="草稿" value="draft" />
            <el-option label="待审批" value="pending" />
            <el-option label="已通过" value="approved" />
            <el-option label="已拒绝" value="rejected" />
            <el-option label="已发布" value="released" />
          </el-select>
        </el-col>
        <el-col :span="10" :offset="10" style="text-align: right">
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>新建发布单
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-card>
      <el-table :data="releases" stripe style="width: 100%">
        <el-table-column prop="id" label="发布单ID" width="100" />
        <el-table-column prop="title" label="标题" min-width="180" />
        <el-table-column prop="created_by" label="提交人" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="180" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button size="small" text @click="viewDetail(row)">查看</el-button>
            <el-button v-if="row.status === 'draft'" size="small" type="primary" text @click="doSubmit(row)">提交审批</el-button>
            <el-button v-if="row.status === 'pending'" size="small" type="success" text @click="doApprove(row)">通过</el-button>
            <el-button v-if="row.status === 'pending'" size="small" type="danger" text @click="doReject(row)">拒绝</el-button>
            <el-button v-if="row.status === 'approved'" size="small" type="warning" text @click="doExecute(row)">发布</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="total > pageSize" style="text-align: center; margin-top: 16px">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="prev, pager, next"
          @current-change="loadReleases"
        />
      </div>
    </el-card>

    <el-dialog v-model="showCreateDialog" title="新建发布单" width="500px">
      <el-form :model="newRelease" label-width="80px">
        <el-form-item label="标题" required>
          <el-input v-model="newRelease.title" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="newRelease.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="doCreate">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showDetailDialog" title="发布单详情" width="700px">
      <template v-if="detail">
        <div><b>标题：</b>{{ detail.title }}</div>
        <div><b>描述：</b>{{ detail.description }}</div>
        <div><b>状态：</b>{{ statusText(detail.status) }}</div>
        <el-divider />
        <div><b>变更项：</b>{{ releaseItems.length }} 项</div>
        <el-table :data="releaseItems" stripe style="width: 100%; margin-top: 8px">
          <el-table-column prop="item_id" label="配置项ID" width="100" />
          <el-table-column label="变更前" width="200">
            <template #default="{ row }"><code>{{ row.value_before }}</code></template>
          </el-table-column>
          <el-table-column label="变更后" width="200">
            <template #default="{ row }"><code>{{ row.value_after }}</code></template>
          </el-table-column>
        </el-table>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  listReleases, createRelease, submitRelease,
  approveRelease, rejectRelease, executeRelease, listReleaseItems,
} from '@/api/release'
import type { ConfigRelease, ConfigReleaseItem } from '@/types'

const releases = ref<ConfigRelease[]>([])
const releaseItems = ref<ConfigReleaseItem[]>([])
const detail = ref<ConfigRelease | null>(null)
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const statusFilter = ref('')
const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const newRelease = ref({ title: '', description: '' })

function statusTag(s: string) {
  const map: Record<string, string> = { draft: 'info', pending: 'warning', approved: 'success', rejected: 'danger', released: 'primary' }
  return map[s] || 'info'
}

function statusText(s: string) {
  const map: Record<string, string> = { draft: '草稿', pending: '待审批', approved: '已通过', rejected: '已拒绝', released: '已发布' }
  return map[s] || s
}

onMounted(() => loadReleases())

async function loadReleases() {
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (statusFilter.value) params.status = statusFilter.value
    const res = await listReleases(params)
    releases.value = res.data || []
    total.value = (res as any).total || 0
  } catch { /* ignore */ }
}

async function doCreate() {
  if (!newRelease.value.title) {
    ElMessage.warning('请输入标题')
    return
  }
  try {
    await createRelease(newRelease.value)
    ElMessage.success('创建成功')
    showCreateDialog.value = false
    newRelease.value = { title: '', description: '' }
    loadReleases()
  } catch { /* ignore */ }
}

async function doSubmit(row: ConfigRelease) {
  try {
    await submitRelease(row.id)
    ElMessage.success('已提交审批')
    loadReleases()
  } catch { /* ignore */ }
}

async function doApprove(row: ConfigRelease) {
  try {
    await ElMessageBox.confirm('确认通过该发布单？')
    await approveRelease(row.id)
    ElMessage.success('已通过')
    loadReleases()
  } catch { /* ignore */ }
}

async function doReject(row: ConfigRelease) {
  try {
    await ElMessageBox.confirm('确认拒绝该发布单？')
    await rejectRelease(row.id)
    ElMessage.success('已拒绝')
    loadReleases()
  } catch { /* ignore */ }
}

async function doExecute(row: ConfigRelease) {
  try {
    await ElMessageBox.confirm('确认发布？发布后不可撤回')
    await executeRelease(row.id)
    ElMessage.success('发布成功')
    loadReleases()
  } catch { /* ignore */ }
}

async function viewDetail(row: ConfigRelease) {
  detail.value = row
  try {
    const res = await listReleaseItems(row.id)
    releaseItems.value = res.data || []
  } catch { /* ignore */ }
  showDetailDialog.value = true
}
</script>
