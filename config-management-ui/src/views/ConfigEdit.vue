<template>
  <div>
    <el-card>
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span>{{ isEdit ? '编辑配置项' : '新建配置项' }}</span>
          <el-button @click="$router.back()">返回</el-button>
        </div>
      </template>

      <el-form :model="form" label-width="120px" style="max-width: 700px">
        <el-form-item label="配置项名称" required>
          <el-input v-model="form.name" placeholder="如：JWT Access Token有效期" />
        </el-form-item>
        <el-form-item label="配置编码" required>
          <el-input v-model="form.code" :disabled="isEdit" placeholder="如：JWT_ACCESS_TOKEN_EXPIRE" />
        </el-form-item>
        <el-form-item label="配置分类" required>
          <el-select v-model="form.group_id" style="width: 100%">
            <el-option v-for="g in groups" :key="g.id" :label="g.name" :value="g.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="数据类型" required>
          <el-select v-model="form.data_type" style="width: 100%">
            <el-option label="STRING - 字符串" value="STRING" />
            <el-option label="INTEGER - 整数" value="INTEGER" />
            <el-option label="BOOLEAN - 布尔" value="BOOLEAN" />
            <el-option label="DECIMAL - 小数" value="DECIMAL" />
            <el-option label="DURATION - 时长" value="DURATION" />
            <el-option label="ENUM - 枚举" value="ENUM" />
            <el-option label="COLOR - 颜色" value="COLOR" />
            <el-option label="CRON - 定时" value="CRON" />
            <el-option label="RATE_LIMIT - 速率限制" value="RATE_LIMIT" />
            <el-option label="LIST - 列表" value="LIST" />
          </el-select>
        </el-form-item>
        <el-form-item label="当前值">
          <el-input v-model="form.current_value" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="默认值">
          <el-input v-model="form.default_value" :disabled="isEdit" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="最小值">
              <el-input v-model="form.min_value" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最大值">
              <el-input v-model="form.max_value" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="可选值">
          <el-input v-model="form.allowed_values" placeholder="逗号分隔，如：true,false" />
        </el-form-item>
        <el-form-item label="配置说明">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="敏感配置">
          <el-switch v-model="form.is_sensitive" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="form.is_enabled" />
        </el-form-item>

        <el-form-item v-if="isEdit" label="变更原因" required>
          <el-input v-model="changeReason" type="textarea" :rows="2" placeholder="请说明变更原因" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSave">
            {{ isEdit ? '保存并提交' : '创建' }}
          </el-button>
          <el-button @click="$router.back()">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { listGroups, getItem, createItem, updateItem } from '@/api/config'
import type { ConfigGroup, ConfigItem } from '@/types'

const route = useRoute()
const router = useRouter()
const isEdit = !!route.params.id

const groups = ref<ConfigGroup[]>([])
const changeReason = ref('')

const form = ref({
  group_id: undefined as number | undefined,
  code: '',
  name: '',
  description: '',
  data_type: 'STRING',
  current_value: '',
  default_value: '',
  min_value: '',
  max_value: '',
  allowed_values: '',
  is_sensitive: false,
  is_enabled: true,
})

onMounted(async () => {
  try {
    const res = await listGroups()
    groups.value = res.data || []
  } catch (e: any) { console.warn('edit failed', e) }

  if (isEdit) {
    try {
      const res = await getItem(Number(route.params.id))
      if (res.data) {
        Object.assign(form.value, res.data)
      }
    } catch (e: any) { console.warn('edit failed', e) }
  }
})

async function handleSave() {
  if (!form.value.name || !form.value.code || !form.value.group_id) {
    ElMessage.warning('请填写必要字段')
    return
  }
  if (isEdit && !changeReason.value) {
    ElMessage.warning('请填写变更原因')
    return
  }

  try {
    if (isEdit) {
      await updateItem(Number(route.params.id), form.value as any, changeReason.value)
      ElMessage.success('保存成功')
    } else {
      await createItem(form.value as any)
      ElMessage.success('创建成功')
    }
    router.push('/config')
  } catch (e: any) { console.warn('edit failed', e) }
}
</script>
