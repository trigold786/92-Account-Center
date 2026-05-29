<template>
  <div>
    <el-tabs v-model="section">
      <el-tab-pane label="角色列表" name="roles" v-if="hasPermission('admin.roles.view')">
        <el-card>
          <template #header>
            <div style="display:flex;justify-content:space-between;align-items:center">
              <span>角色列表</span>
              <el-button type="primary" @click="showCreateDialog = true" v-if="hasPermission('admin.roles.create')">新建角色</el-button>
            </div>
          </template>
          <el-table :data="roles" stripe>
            <el-table-column prop="id" label="ID" width="60" />
            <el-table-column prop="name" label="名称" width="150" />
            <el-table-column prop="description" label="描述" />
            <el-table-column label="权限数" width="80">
              <template #default="{ row }">{{ permCounts[row.id] ?? '...' }}</template>
            </el-table-column>
            <el-table-column label="操作" width="160">
              <template #default="{ row }">
                <el-button size="small" @click="editRole(row)" v-if="hasPermission('admin.roles.edit')">编辑</el-button>
                <el-button size="small" type="danger" @click="deleteRole(row)" v-if="hasPermission('admin.roles.delete')">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="权限配置" name="perms" v-if="hasPermission('admin.roles.permission')">
        <el-card>
          <template #header>
            <div style="display:flex;gap:12px;align-items:center">
              <span>选择角色：</span>
              <el-select v-model="selectedRoleId" placeholder="请选择角色" @change="loadRolePerms" style="width:200px">
                <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
              </el-select>
            </div>
          </template>
          <div v-if="selectedRoleId">
            <div v-for="(perms, group) in permissionGroups" :key="group" style="margin-bottom:16px">
              <h4 style="margin:0 0 8px;color:var(--el-text-color-secondary)">{{ group }}</h4>
              <el-checkbox-group v-model="selectedPerms">
                <el-checkbox v-for="p in perms" :key="p" :label="p" :value="p" border style="margin:4px 8px 4px 0">
                  {{ p.split('.').slice(1).join('.') }}
                </el-checkbox>
              </el-checkbox-group>
            </div>
            <el-button type="primary" @click="savePermissions">保存权限</el-button>
            <el-button @click="loadRolePerms">重置</el-button>
          </div>
          <div v-else style="color:var(--el-text-color-placeholder);padding:40px;text-align:center">请先选择一个角色</div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="用户角色" name="users" v-if="hasPermission('admin.roles.assign')">
        <el-card>
          <template #header>
            <div style="display:flex;gap:12px;align-items:center">
              <span>用户账户 ID：</span>
              <el-input v-model="targetUserId" placeholder="如 admin_user / normal_user" style="width:240px" @keyup.enter="loadUserRoles" />
              <el-button type="primary" @click="loadUserRoles">查询</el-button>
            </div>
          </template>
          <div v-if="targetUserId && userRoleList !== null">
            <h4>当前角色</h4>
            <div style="display:flex;flex-wrap:wrap;gap:8px;margin-bottom:16px">
              <el-tag v-for="ur in userRoleList" :key="ur.role_id" closable @close="removeUserRole(ur.role_id)">
                {{ roleName(ur.role_id) }}
              </el-tag>
              <span v-if="userRoleList.length === 0" style="color:var(--el-text-color-placeholder)">暂无角色</span>
            </div>
            <h4>添加角色</h4>
            <div style="display:flex;gap:8px">
              <el-select v-model="addRoleId" placeholder="选择角色" style="width:200px">
                <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id" />
              </el-select>
              <el-button type="primary" @click="assignUserRole">添加</el-button>
            </div>
          </div>
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="showCreateDialog" :title="editingRole ? '编辑角色' : '新建角色'" width="400px">
      <el-form label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="roleForm.name" placeholder="角色标识" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="roleForm.description" placeholder="角色描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveRole">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { usePermissionStore } from '@/store/permission'

const perm = usePermissionStore()
const hasPermission = (p: string) => perm.hasPermission(p)
import {
  listRoles, createRole, deleteRole as apiDeleteRole, getRolePermissions, addRolePermission, removeRolePermission,
  getUserRoles, setUserRole, removeUserRole,
} from '@/api/roles'

const section = ref('roles')
const roles = ref<any[]>([])
const permCounts = reactive<Record<number, number>>({})

onMounted(async () => {
  await loadRoles()
})

async function loadRoles() {
  try {
    const r = await listRoles()
    roles.value = (r.data.data || [])
    for (const role of roles.value) {
      try {
        const pr = await getRolePermissions(role.id)
        permCounts[role.id] = (pr.data.data || []).length
      } catch { permCounts[role.id] = 0 }
    }
  } catch { ElMessage.error('加载角色列表失败') }
}

// role CRUD
const showCreateDialog = ref(false)
const editingRole = ref<any>(null)
const roleForm = reactive({ name: '', description: '' })

function editRole(role: any) {
  editingRole.value = role
  roleForm.name = role.name
  roleForm.description = role.description
  showCreateDialog.value = true
}

async function saveRole() {
  if (!roleForm.name) { ElMessage.warning('请输入角色名称'); return }
  try {
    await createRole({ name: roleForm.name, description: roleForm.description })
    ElMessage.success('保存成功')
    showCreateDialog.value = false
    roleForm.name = ''; roleForm.description = ''
    editingRole.value = null
    await loadRoles()
  } catch (e: any) { ElMessage.error(e.message) }
}

async function deleteRole(role: any) {
  try {
    await ElMessageBox.confirm(`确定删除角色 "${role.name}"？此操作不可撤销。`, '确认')
    await apiDeleteRole(role.id)
    ElMessage.success('角色已删除')
    await loadRoles()
  } catch {}
}

// permission config
const selectedRoleId = ref<number | null>(null)
const selectedPerms = ref<string[]>([])
const allRolePerms = ref<any[]>([])

const permissionGroups: Record<string, string[]> = {
  '导航菜单': ['nav.dashboard', 'nav.account', 'nav.credits', 'nav.subscriptions', 'nav.referral', 'nav.devices', 'nav.admin'],
  '页面访问': ['page.dashboard', 'page.account', 'page.credits', 'page.subscriptions', 'page.referral', 'page.devices', 'page.admin'],
  '账户-修改密码': ['account.password.change'],
  '账户-注销': ['account.delete.apply', 'account.delete.cancel'],
  '推荐-复制链接': ['referral.copy'],
  '设备-信任/移除': ['device.trust', 'device.remove'],
  '管理后台-概览': ['admin.overview.view'],
  '管理后台-风险': ['admin.risk.view'],
  '管理后台-黑名单': ['admin.blacklist.view', 'admin.blacklist.add', 'admin.blacklist.delete'],
  '管理后台-审计': ['admin.audit.view', 'admin.audit.verify'],
  '管理后台-SMS': ['admin.sms.view'],
  '管理后台-角色管理': ['admin.roles.view', 'admin.roles.create', 'admin.roles.edit', 'admin.roles.delete'],
  '管理后台-权限配置': ['admin.roles.permission'],
  '管理后台-用户角色分配': ['admin.roles.assign'],
  '用户管理': ['admin.user.manage', 'admin.user.freeze', 'admin.user.ban'],
  '积分管理(后台)': ['admin.credit.adjust', 'admin.plan.manage', 'admin.coupon.manage'],
  '审核/黑名单(后台)': ['admin.audit.view', 'admin.blacklist.manage'],
  '配置管理': ['config.read', 'config.edit', 'config.delete'],
  '发布管理': ['release.create', 'release.approve', 'release.execute'],
  '财务': ['finance.order.view', 'finance.refund.approve', 'finance.invoice.manage'],
  '数据': ['data.dashboard', 'data.rfm', 'data.funnel'],
  'SMS(后台)': ['sms.status'],
  '自助服务': ['account.self', 'credits.self', 'subscriptions.self', 'devices.self', 'referral.self', 'data.rfm.self'],
  '审计/权限(后台)': ['audit.view', 'permission.manage'],
}

async function loadRolePerms() {
  if (!selectedRoleId.value) return
  try {
    const r = await getRolePermissions(selectedRoleId.value)
    allRolePerms.value = r.data.data || []
    selectedPerms.value = allRolePerms.value.map((p: any) => p.permission)
  } catch { ElMessage.error('加载权限失败') }
}

async function savePermissions() {
  if (!selectedRoleId.value) return
  const current = new Set(allRolePerms.value.map((p: any) => p.permission))
  const desired = new Set(selectedPerms.value)

  const toAdd = selectedPerms.value.filter(p => !current.has(p))
  const toRemove = allRolePerms.value.filter((p: any) => !desired.has(p.permission))

  try {
    for (const p of toAdd) {
      await addRolePermission(selectedRoleId.value, p)
    }
    for (const p of toRemove) {
      await removeRolePermission(selectedRoleId.value, p.id)
    }
    ElMessage.success('权限已更新')
    await loadRolePerms()
    await loadRoles()
  } catch (e: any) { ElMessage.error(e.message) }
}

// user role assignment
const targetUserId = ref('')
const userRoleList = ref<any[] | null>(null)
const addRoleId = ref<number | null>(null)

async function loadUserRoles() {
  if (!targetUserId.value) return
  try {
    const r = await getUserRoles(targetUserId.value)
    userRoleList.value = r.data.data || []
  } catch { ElMessage.error('查询用户角色失败') }
}

function roleName(roleId: number) {
  return roles.value.find(r => r.id === roleId)?.name || `role_${roleId}`
}

async function assignUserRole() {
  if (!targetUserId.value || !addRoleId.value) return
  try {
    await setUserRole(targetUserId.value, addRoleId.value)
    ElMessage.success('角色已分配')
    addRoleId.value = null
    await loadUserRoles()
  } catch (e: any) { ElMessage.error(e.message) }
}

async function removeUserRole(roleId: number) {
  if (!targetUserId.value) return
  try {
    await removeUserRole(targetUserId.value, roleId)
    ElMessage.success('角色已移除')
    await loadUserRoles()
  } catch (e: any) { ElMessage.error(e.message) }
}
</script>
