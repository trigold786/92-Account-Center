<template>
  <div>
    <el-card><template #header>账户信息</template>
      <p>用户ID: {{ auth.userId }}</p>
      <p>账号: {{ auth.accountId }}</p>
      <p>等级: {{ tier }}</p>
    </el-card>

    <el-card style="margin-top:16px"><template #header>修改密码</template>
      <el-form label-width="100px">
        <el-form-item label="验证方式"><el-input v-model="pwd.credential" placeholder="手机号/邮箱" /></el-form-item>
        <el-form-item><el-button @click="sendPwdCode" :disabled="pwdSending">{{ pwdBtnText }}</el-button></el-form-item>
        <el-form-item label="验证码"><el-input v-model="pwd.code" /></el-form-item>
        <el-form-item label="新密码"><el-input v-model="pwd.newPassword" type="password" /></el-form-item>
        <el-form-item><el-button type="primary" @click="changePwd">确认修改</el-button></el-form-item>
      </el-form>
    </el-card>

    <el-card style="margin-top:16px"><template #header>注销账户</template>
      <p style="color:var(--text-secondary);margin-bottom:12px">注销后将有30天冻结期，期间可撤销</p>
      <el-button type="danger" plain :loading="delLoading" @click="requestDel">{{ delStatus ? '查看注销状态' : '申请注销' }}</el-button>
      <el-button v-if="delStatus" type="info" @click="cancelDel">撤销申请</el-button>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useAuthStore } from '@/store/auth'
import { getTier, changePassword, sendPasswordCode, requestDeletion, cancelDeletion, getDeletionStatus } from '@/api/account'
import { ElMessage } from 'element-plus'

const auth = useAuthStore()
const tier = ref('')
const pwd = reactive({ credential: '', code: '', newPassword: '' })
const pwdSending = ref(false); const pwdBtnText = ref('发送验证码')
const delLoading = ref(false); const delStatus = ref(false)

onMounted(async () => {
  try { const r = await getTier(auth.userId); tier.value = r.data.data?.tier } catch {}
  try { const r = await getDeletionStatus(); delStatus.value = r.data.data?.status } catch {}
})

async function sendPwdCode() {
  if (!pwd.credential) { ElMessage.warning('请输入联系方式'); return }
  pwdSending.value = true
  try { await sendPasswordCode(pwd.credential); ElMessage.success('验证码已发送'); let s = 60; pwdBtnText.value = `${s}s`; const t = setInterval(() => { s--; pwdBtnText.value = `${s}s`; if (s <= 0) { clearInterval(t); pwdBtnText.value = '重新获取'; pwdSending.value = false } }, 1000) } catch { pwdSending.value = false }
}

async function changePwd() {
  if (!pwd.code || !pwd.newPassword) { ElMessage.warning('请填写完整'); return }
  try { await changePassword({ current_password: pwd.credential, new_password: pwd.newPassword, code: pwd.code }); ElMessage.success('密码修改成功'); pwd.credential = ''; pwd.code = ''; pwd.newPassword = '' } catch (e: any) { ElMessage.error(e.message) }
}

async function requestDel() {
  delLoading.value = true
  try { await requestDeletion(); ElMessage.success('注销申请已提交'); delStatus.value = true } catch (e: any) { ElMessage.error(e.message) }
  delLoading.value = false
}

async function cancelDel() {
  try { await cancelDeletion(); ElMessage.success('已撤销注销'); delStatus.value = false } catch (e: any) { ElMessage.error(e.message) }
}
</script>
