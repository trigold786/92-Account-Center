import { register, sendVerificationCode } from '../../utils/auth'
import { loginWithPassword, loginWithCode } from '../../utils/auth'

Page({
  data: { phone: '', password: '', confirmPassword: '', code: '', codeBtnText: '获取验证码', codeBtnDisabled: false, loading: false, agreeTerms: false },
  onPhoneInput(e: any) { this.setData({ phone: e.detail.value }) },
  onPasswordInput(e: any) { this.setData({ password: e.detail.value }) },
  onConfirmInput(e: any) { this.setData({ confirmPassword: e.detail.value }) },
  onCodeInput(e: any) { this.setData({ code: e.detail.value }) },
  toggleAgree() { this.setData({ agreeTerms: !this.data.agreeTerms }) },
  async sendCode() {
    const phone = this.data.phone
    if (!/^1\d{10}$/.test(phone)) { wx.showToast({ title: '请输入正确手机号', icon: 'none' }); return }
    this.setData({ codeBtnDisabled: true })
    try {
      await sendVerificationCode(phone, 'register')
      wx.showToast({ title: '验证码已发送', icon: 'success' })
      let sec = 60
      const timer = setInterval(() => {
        sec--
        this.setData({ codeBtnText: `${sec}s` })
        if (sec <= 0) { clearInterval(timer); this.setData({ codeBtnText: '重新获取', codeBtnDisabled: false }) }
      }, 1000)
    } catch (e: any) { wx.showToast({ title: e.message || '发送失败', icon: 'none' }); this.setData({ codeBtnDisabled: false }) }
  },
  async doRegister() {
    if (this.data.loading) return
    const { phone, password, confirmPassword, code, agreeTerms } = this.data
    if (!phone || !password || !code) { wx.showToast({ title: '请填写完整信息', icon: 'none' }); return }
    if (password.length < 6) { wx.showToast({ title: '密码至少6位', icon: 'none' }); return }
    if (password !== confirmPassword) { wx.showToast({ title: '两次密码不一致', icon: 'none' }); return }
    if (!agreeTerms) { wx.showToast({ title: '请同意服务条款', icon: 'none' }); return }
    this.setData({ loading: true })
    try {
      await register({ phone, password, code })
      await loginWithPassword(phone, password)
      wx.showToast({ title: '注册成功', icon: 'success' })
      setTimeout(() => wx.switchTab({ url: '/pages/index/index' }), 1500)
    } catch (e: any) { wx.showToast({ title: e.message || '注册失败', icon: 'none' }) }
    this.setData({ loading: false })
  },
  goLogin() { wx.navigateBack() },
})
