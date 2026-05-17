<template>
  <div>
    <el-card style="margin-bottom: 16px">
      <el-row :gutter="16">
        <el-col :span="6">
          <el-input v-model="search.code" placeholder="搜索配置项" clearable @input="handleSearch" />
        </el-col>
        <el-col :span="4">
          <el-select v-model="search.group_id" placeholder="按服务筛选" clearable @change="handleSearch" style="width: 100%">
            <el-option v-for="g in groups" :key="g.id" :label="g.name" :value="g.id" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="search.data_type" placeholder="按类型筛选" clearable @change="handleSearch" style="width: 100%">
            <el-option label="STRING" value="STRING" />
            <el-option label="INTEGER" value="INTEGER" />
            <el-option label="BOOLEAN" value="BOOLEAN" />
            <el-option label="DECIMAL" value="DECIMAL" />
            <el-option label="DURATION" value="DURATION" />
            <el-option label="ENUM" value="ENUM" />
            <el-option label="COLOR" value="COLOR" />
            <el-option label="CRON" value="CRON" />
            <el-option label="RATE_LIMIT" value="RATE_LIMIT" />
            <el-option label="LIST" value="LIST" />
          </el-select>
        </el-col>
        <el-col :span="6" style="text-align: right">
          <el-button type="primary" @click="$router.push('/config/edit')">
            <el-icon><Plus /></el-icon>添加配置
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-row :gutter="16">
      <el-col :span="5">
        <el-card>
          <template #header><span>配置分组</span></template>
          <el-tree
            :data="groupTree"
            :props="{ label: 'name', children: 'children' }"
            node-key="id"
            default-expand-all
            highlight-current
            @node-click="onGroupClick"
          />
        </el-card>
      </el-col>
      <el-col :span="19">
        <el-card>
          <div v-for="item in items" :key="item.id" class="config-card" @click="editItem(item)">
            <div class="config-header">
              <span class="config-name">{{ item.name }}</span>
              <el-tooltip :content="item.description || item.code" placement="top">
                <el-icon class="help-icon"><QuestionFilled /></el-icon>
              </el-tooltip>
              <el-tag v-if="item.is_sensitive" type="danger" size="small" style="margin-left: 8px">敏感</el-tag>
              <el-tag :type="item.is_enabled ? 'success' : 'info'" size="small" style="margin-left: 4px">
                {{ item.is_enabled ? '启用' : '禁用' }}
              </el-tag>
            </div>
            <div class="config-code">{{ item.code }}</div>
            <div class="config-value">
              <code>{{ item.is_sensitive ? '***' : item.current_value }}</code>
              <span class="config-type">{{ item.data_type }}</span>
            </div>
            <div class="config-actions">
              <el-button size="small" @click.stop="editItem(item)">编辑</el-button>
              <el-button size="small" @click.stop="viewVersions(item)">历史</el-button>
              <el-button size="small" type="danger" plain @click.stop="deleteItem(item)">删除</el-button>
              <el-button size="small" @click.stop="resetItem(item)">重置</el-button>
            </div>
          </div>

          <el-empty v-if="items.length === 0" description="暂无配置项" />

          <div v-if="total > pageSize" style="text-align: center; margin-top: 16px">
            <el-pagination
              v-model:current-page="page"
              :page-size="pageSize"
              :total="total"
              layout="prev, pager, next"
              @current-change="loadItems"
            />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="versionDialog" title="版本历史" width="600px">
      <el-timeline>
        <el-timeline-item
          v-for="v in versions"
          :key="v.id"
          :timestamp="v.created_at"
          placement="top"
        >
          <div><b>变更人：</b>{{ v.changed_by }}</div>
          <div><b>变更原因：</b>{{ v.change_reason }}</div>
          <div><b>变更前：</b><code>{{ v.value_before }}</code></div>
          <div><b>变更后：</b><code>{{ v.value_after }}</code></div>
        </el-timeline-item>
      </el-timeline>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listGroups, listItems, listVersions, deleteItem as deleteConfigItem, resetItemToDefault } from '@/api/config'
import type { ConfigGroup, ConfigItem, ConfigVersion } from '@/types'

const router = useRouter()
const groups = ref<ConfigGroup[]>([])
const groupTree = ref<any[]>([])
const items = ref<ConfigItem[]>([])
const versions = ref<ConfigVersion[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const versionDialog = ref(false)

const search = ref({ code: '', group_id: undefined as number | undefined, data_type: '' })

onMounted(async () => {
  await loadGroups()
  await loadItems()
})

async function loadGroups() {
  try {
    const res = await listGroups()
    groups.value = res.data || []
    groupTree.value = [
      { id: -1, name: '全部', children: res.data.map((g) => ({ id: g.id, name: g.name })) },
    ]
  } catch (e: any) { console.warn('load failed', e) }
}

async function loadItems() {
  try {
    const params: any = { page: page.value, page_size: pageSize.value }
    if (search.value.code) params.code = search.value.code
    if (search.value.group_id) params.group_id = search.value.group_id
    if (search.value.data_type) params.data_type = search.value.data_type
    const res = await listItems(params)
    items.value = res.data || []
    total.value = (res as any).total || 0
  } catch (e: any) { console.warn('load items failed', e) }
}

function handleSearch() {
  page.value = 1
  loadItems()
}

function onGroupClick(data: any) {
  if (data.id === -1 || data.children) return
  search.value.group_id = data.id
  handleSearch()
}

function editItem(item: ConfigItem) {
  router.push(`/config/edit/${item.id}`)
}

async function viewVersions(item: ConfigItem) {
  try {
    const res = await listVersions(item.id)
    versions.value = res.data || []
    versionDialog.value = true
  } catch (e: any) { console.warn('load versions failed', e) }
}

async function deleteItem(item: ConfigItem) {
  try {
    await ElMessageBox.confirm(`确定要删除配置项 "${item.code}" 吗？`, '确认删除', {
      type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消'
    })
    await deleteConfigItem(item.id)
    ElMessage.success('删除成功')
    await loadItems()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error('删除失败: ' + (e.message || e))
  }
}

async function resetItem(item: ConfigItem) {
  try {
    await ElMessageBox.confirm(`确定要将 "${item.code}" 重置为默认值吗？`, '确认重置', {
      type: 'info', confirmButtonText: '重置', cancelButtonText: '取消'
    })
    await resetItemToDefault(item.id)
    ElMessage.success('重置成功')
    await loadItems()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error('重置失败: ' + (e.message || e))
  }
}
</script>

<style scoped>
.config-card {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 12px 16px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: box-shadow 0.2s;
}
.config-card:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}
.config-header {
  display: flex;
  align-items: center;
}
.config-name {
  font-weight: 600;
  font-size: 14px;
}
.help-icon {
  margin-left: 4px;
  color: #c0c4cc;
  cursor: help;
}
.config-code {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}
.config-value {
  margin-top: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.config-value code {
  background: #f5f7fa;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 14px;
}
.config-type {
  font-size: 11px;
  color: #909399;
  background: #f0f0f0;
  padding: 1px 6px;
  border-radius: 3px;
}
.config-actions {
  margin-top: 8px;
  display: flex;
  gap: 8px;
}
</style>
