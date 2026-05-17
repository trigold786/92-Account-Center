<template>
  <div>
    <el-row :gutter="16">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div style="display: flex; justify-content: space-between">
              <span>角色管理</span>
              <el-button size="small" type="primary" @click="showRoleDialog = true">
                <el-icon><Plus /></el-icon>新建角色
              </el-button>
            </div>
          </template>
          <el-table :data="roles" stripe @row-click="selectRole">
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="name" label="角色名称" />
            <el-table-column prop="description" label="描述" />
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card v-if="selectedRole">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span>权限配置 - {{ selectedRole.name }}</span>
              <el-button size="small" type="primary" @click="showAddPermissionDialog = true">
                <el-icon><Plus /></el-icon>添加权限
              </el-button>
            </div>
          </template>
          <el-table :data="permissions" stripe>
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="permission" label="权限" />
          </el-table>
        </el-card>

        <el-card v-else>
          <el-empty description="请选择一个角色" />
        </el-card>

        <el-dialog v-model="showAddPermissionDialog" title="添加权限" width="400px">
          <el-form :model="newPermission" label-width="80px">
            <el-form-item label="权限名称" required>
              <el-input v-model="newPermission.permission" placeholder="输入权限标识" />
            </el-form-item>
          </el-form>
          <template #footer>
            <el-button @click="showAddPermissionDialog = false">取消</el-button>
            <el-button type="primary" @click="doAddPermission">添加</el-button>
          </template>
        </el-dialog>
      </el-col>
    </el-row>

    <el-dialog v-model="showRoleDialog" title="新建角色" width="400px">
      <el-form :model="newRole" label-width="80px">
        <el-form-item label="名称" required>
          <el-input v-model="newRole.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="newRole.description" type="textarea" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRoleDialog = false">取消</el-button>
        <el-button type="primary" @click="doCreateRole">创建</el-button>
      </template>
    </el-dialog>

    <el-card style="margin-top: 16px">
      <template #header>
        <span>用户角色分配</span>
      </template>
      <el-row :gutter="16">
        <el-col :span="6">
          <el-input v-model="userSearch" placeholder="输入用户ID" />
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="searchUserRoles">查询</el-button>
        </el-col>
      </el-row>
      <el-table v-if="userRoles.length > 0" :data="userRoles" stripe style="margin-top: 8px">
        <el-table-column prop="user_id" label="用户ID" />
        <el-table-column label="角色">
          <template #default="{ row }">
            {{ roleName(row.role_id) }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listRoles, createRole, getRolePermissions, addRolePermission } from '@/api/permission'
import { getUserRoles, setUserRole } from '@/api/permission'
import type { Role, RolePermission, UserRole } from '@/types'

const roles = ref<Role[]>([])
const permissions = ref<RolePermission[]>([])
const userRoles = ref<UserRole[]>([])
const selectedRole = ref<Role | null>(null)
const showRoleDialog = ref(false)
const userSearch = ref('')

const newRole = ref({ name: '', description: '' })
const showAddPermissionDialog = ref(false)
const newPermission = ref({ permission: '' })

onMounted(() => loadRoles())

async function loadRoles() {
  try {
    const res = await listRoles()
    roles.value = res.data || []
  } catch (e: any) { console.warn('permission operation failed', e) }
}

async function selectRole(role: Role) {
  selectedRole.value = role
  try {
    const res = await getRolePermissions(role.id)
    permissions.value = res.data || []
  } catch (e: any) { console.warn('permission operation failed', e) }
}

function roleName(id: number) {
  return roles.value.find((r) => r.id === id)?.name || `角色#${id}`
}

async function doCreateRole() {
  if (!newRole.value.name) {
    ElMessage.warning('请输入角色名称')
    return
  }
  try {
    await createRole(newRole.value)
    ElMessage.success('创建成功')
    showRoleDialog.value = false
    newRole.value = { name: '', description: '' }
    loadRoles()
  } catch (e: any) { console.warn('permission operation failed', e) }
}

async function searchUserRoles() {
  if (!userSearch.value) return
  try {
    const res = await getUserRoles(userSearch.value)
    userRoles.value = res.data || []
  } catch (e: any) { console.warn('permission operation failed', e) }
}

async function doAddPermission() {
  if (!newPermission.value.permission || !selectedRole.value) {
    ElMessage.warning('请输入权限名称')
    return
  }
  try {
    await addRolePermission(selectedRole.value.id, { permission: newPermission.value.permission })
    ElMessage.success('权限添加成功')
    showAddPermissionDialog.value = false
    newPermission.value = { permission: '' }
    const res = await getRolePermissions(selectedRole.value.id)
    permissions.value = res.data || []
  } catch (e: any) { console.warn('permission operation failed', e) }
}
</script>
