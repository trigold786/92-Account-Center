import { loginWithPassword, loginWithCode, sendSMSCode } from '../../utils/auth'
import { setUserId } from '../../utils/storage'

Page({
  data: {
    phone: '',
    password: '',
    code: '',
    loginMode: 'password' as 'password' | 'code',
    codeBtnText: '获取验证码',
    codeBtnDisabled: false,
    loading: false,
  },
  switchMode() {
    this.setData({
      loginMode: this.data.loginMode === 'password' ? 'code' : 'password',
      password: '',
      code: '',
    })
  },
  onPhoneInput(e: any) { this.setData({ phone: e.detail.value }) },
  onPasswordInput(e: any) { this.setData({ password: e.detail.value }) },
  onCodeInput(e: any) { this.setData({ code: e.detail.value }) },
  async sendCode() {
    const phone = this.data.phone
    if (!/^1\d{10}$/.test(phone)) { wx.showToast({ title: '请输入正确手机号', icon: 'none' }); return }
    this.setData({ codeBtnDisabled: true })
    try {
      await sendSMSCode(phone)
      wx.showToast({ title: '验证码已发送', icon: 'success' })
      let sec = 60
      const timer = setInterval(() => {
        sec--
        this.setData({ codeBtnText: `${sec}s` })
        if (sec <= 0) { clearInterval(timer); this.setData({ codeBtnText: '重新获取', codeBtnDisabled: false }) }
      }, 1000)
    } catch (e: any) {
      wx.showToast({ title: e.message || '发送失败', icon: 'none' })
      this.setData({ codeBtnDisabled: false })
    }
  },
  async doLogin() {
    if (this.data.loading) return
    if (!this.data.phone) { wx.showToast({ title: '请输入手机号', icon: 'none' }); return }
    this.setData({ loading: true })
    try {
      if (this.data.loginMode === 'password') {
        if (!this.data.password) { wx.showToast({ title: '请输入密码', icon: 'none' }); this.setData({ loading: false }); return }
        await loginWithPassword(this.data.phone, this.data.password)
      } else {
        if (!this.data.code) { wx.showToast({ title: '请输入验证码', icon: 'none' }); this.setData({ loading: false }); return }
        await loginWithCode(this.data.phone, this.data.code)
      }
      wx.showToast({ title: '登录成功', icon: 'success' })
      setTimeout(() => wx.switchTab({ url: '/pages/index/index' }), 1500)
    } catch (e: any) {
      wx.showToast({ title: e.message || '登录失败', icon: 'none' })
    }
    this.setData({ loading: false })
  },
  goRegister() { wx.navigateTo({ url: '/pages/register/register' }) },
  onWechatLogin() {
    wx.showToast({ title: '微信登录开发中', icon: 'none' })
  },
})
