<template>
  <div>
    <el-card><template #header>账户信息</template>
      <p>用户ID: {{ auth.userId }}</p>
      <p>账号: {{ auth.accountId }}</p>
      <p>等级: {{ tier }}</p>
    </el-card>

    <el-card style="margin-top:16px" v-if="hasPermission('account.password.change')"><template #header>修改密码</template>
      <el-form label-width="100px">
        <el-form-item label="验证方式">
          <el-select v-model="pwd.verificationType" placeholder="选择验证方式">
            <el-option label="短信验证码" value="sms_code" />
            <el-option label="邮箱 OTP" value="email_otp" />
            <el-option label="当前密码" value="password" />
          </el-select>
        </el-form-item>
        <el-form-item v-if="pwd.verificationType !== 'password'" label="联系方式">
          <el-input v-model="pwd.credential" :placeholder="pwd.verificationType === 'email_otp' ? '邮箱地址' : '手机号'" @keyup.enter="sendPwdCode" />
        </el-form-item>
        <el-form-item v-if="pwd.verificationType !== 'password'">
          <el-button @click="sendPwdCode" :disabled="pwdSending">{{ pwdBtnText }}</el-button>
        </el-form-item>
        <el-form-item v-if="pwd.verificationType !== 'password'" label="验证码"><el-input v-model="pwd.verificationCode" @keyup.enter="changePwd" /></el-form-item>
        <el-form-item v-if="pwd.verificationType === 'password'" label="当前密码"><el-input v-model="pwd.currentPassword" type="password" @keyup.enter="changePwd" /></el-form-item>
        <el-form-item label="新密码"><el-input v-model="pwd.newPassword" type="password" @keyup.enter="changePwd" /></el-form-item>
        <el-form-item><el-button type="primary" @click="changePwd">确认修改</el-button></el-form-item>
      </el-form>
    </el-card>

    <el-card style="margin-top:16px" v-if="hasPermission('account.delete.apply')"><template #header>注销账户</template>
      <p style="color:var(--text-secondary);margin-bottom:12px">注销后将有30天冻结期，期间可撤销</p>
      <div v-if="!delStatus" style="margin-bottom:12px">
        <el-form label-width="100px">
          <el-form-item label="验证方式">
            <el-select v-model="del.verificationType">
              <el-option label="短信验证码" value="sms_code" />
              <el-option label="邮箱 OTP" value="email_otp" />
            </el-select>
          </el-form-item>
          <el-form-item label="验证码"><el-input v-model="del.verificationCode" placeholder="输入验证码" /></el-form-item>
        </el-form>
      </div>
      <el-button type="danger" plain :loading="delLoading" @click="requestDel">{{ delStatus ? '查看注销状态' : '申请注销' }}</el-button>
      <el-button v-if="delStatus" type="info" @click="cancelDel">撤销申请</el-button>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useAuthStore } from '@/store/auth'
import { usePermissionStore } from '@/store/permission'
import { getTier, changePassword, sendPasswordCode, requestDeletion, cancelDeletion, getDeletionStatus } from '@/api/account'
import { sendSMSCode } from '@/api/auth'
import { ElMessage } from 'element-plus'

const auth = useAuthStore()
const perm = usePermissionStore()
const hasPermission = (p: string) => perm.hasPermission(p)
const tier = ref('')
const pwd = reactive({ credential: '', verificationCode: '', newPassword: '', currentPassword: '', verificationType: 'sms_code' })
const pwdSending = ref(false); const pwdBtnText = ref('发送验证码')
const delLoading = ref(false); const delStatus = ref(false)
const del = reactive({ verificationCode: '', verificationType: 'sms_code' })

onMounted(async () => {
  try { const r = await getTier(auth.userId); tier.value = r.data.data?.tier ?? r.data.identity_tier } catch {}
  try { const r = await getDeletionStatus(); delStatus.value = !!(r.data.data?.requested_at ?? r.data.requested_at) } catch {}
})

async function sendPwdCode() {
  if (pwd.verificationType === 'password') return
  if (!pwd.credential) { ElMessage.warning('请输入联系方式'); return }
  pwdSending.value = true
  try {
    if (pwd.verificationType === 'sms_code') {
      await sendSMSCode(pwd.credential)
    } else {
      await sendPasswordCode(pwd.credential)
    }
    const isDev = import.meta.env.DEV
    ElMessage.success(isDev ? '验证码已发送（开发模式验证码: 012345）' : '验证码已发送')
    let s = 60; pwdBtnText.value = `${s}s`
    const t = setInterval(() => { s--; pwdBtnText.value = `${s}s`; if (s <= 0) { clearInterval(t); pwdBtnText.value = '重新获取'; pwdSending.value = false } }, 1000)
  } catch (e: any) {
    const msg = e?.response?.data?.error || e?.response?.data?.message || '发送失败'
    if (msg.includes('rate limit') || msg.includes('频繁')) {
      ElMessage.warning('发送过于频繁，请稍后再试')
    } else {
      ElMessage.error(msg)
    }
    pwdSending.value = false
  }
}

async function changePwd() {
  if (!pwd.newPassword) { ElMessage.warning('请输入新密码'); return }
  if (pwd.verificationType !== 'password' && !pwd.verificationCode) { ElMessage.warning('请输入验证码'); return }
  if (pwd.verificationType === 'password' && !pwd.currentPassword) { ElMessage.warning('请输入当前密码'); return }
  try {
    await changePassword({
      current_password: pwd.currentPassword,
      new_password: pwd.newPassword,
      verification_code: pwd.verificationCode,
      verification_type: pwd.verificationType,
    })
    ElMessage.success('密码修改成功')
    pwd.credential = ''; pwd.verificationCode = ''; pwd.newPassword = ''; pwd.currentPassword = ''
  } catch (e: any) { ElMessage.error(e.message) }
}

async function requestDel() {
  if (delStatus) {
    try { const r = await getDeletionStatus(); ElMessage.info(`注销状态: ${JSON.stringify(r.data)}`) } catch (e: any) { ElMessage.error(e.message) }
    return
  }
  if (!del.verificationCode) { ElMessage.warning('请输入验证码'); return }
  delLoading.value = true
  try {
    await requestDeletion({ verification_code: del.verificationCode, verification_type: del.verificationType })
    ElMessage.success('注销申请已提交'); delStatus.value = true
  } catch (e: any) { ElMessage.error(e.message) }
  delLoading.value = false
}

async function cancelDel() {
  try { await cancelDeletion(); ElMessage.success('已撤销注销'); delStatus.value = false } catch (e: any) { ElMessage.error(e.message) }
}
</script>
