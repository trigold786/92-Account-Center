import { api } from '../../utils/api'
import { getUserId, isLoggedIn, clearAuth } from '../../utils/storage'
import type { DeviceInfo, RiskEvent } from '../../models/types'

Page({
  data: {
    devices: [] as DeviceInfo[],
    riskEvents: [] as RiskEvent[],
    loading: true,
    currentPassword: '',
    newPassword: '',
    confirmPassword: '',
    passMsg: '',
    passSuccess: false,
  },
  onShow() {
    if (!isLoggedIn()) { wx.showToast({ title: '请先登录', icon: 'none' }); setTimeout(() => wx.navigateTo({ url: '/pages/login/login' }), 500); return }
    this.loadData()
  },
  async loadData() {
    this.setData({ loading: true })
    const uid = getUserId()
    const [devRes, riskRes] = await Promise.all([
      api.get<DeviceInfo[]>(`/device/user/${uid}`),
      api.get<RiskEvent[]>(`/risk/history/${uid}`),
    ])
    this.setData({ devices: devRes.data || [], riskEvents: (riskRes.data || []).slice(0, 10), loading: false })
  },
  onCurrentPwd(e: any) { this.setData({ currentPassword: e.detail.value }) },
  onNewPwd(e: any) { this.setData({ newPassword: e.detail.value }) },
  onConfirmPwd(e: any) { this.setData({ confirmPassword: e.detail.value }) },
  async changePassword() {
    const { currentPassword, newPassword, confirmPassword } = this.data
    if (!currentPassword || !newPassword) { this.setData({ passMsg: '请填写完整', passSuccess: false }); return }
    if (newPassword.length < 6) { this.setData({ passMsg: '密码至少6位', passSuccess: false }); return }
    if (newPassword !== confirmPassword) { this.setData({ passMsg: '两次密码不一致', passSuccess: false }); return }
    wx.showLoading({ title: '处理中...' })
    try {
      const res = await api.post('/account/password/send-verification-code', { credential: '' })
      if (res.code === 0) {
        this.setData({ passMsg: '密码修改请求已提交', passSuccess: true, currentPassword: '', newPassword: '', confirmPassword: '' })
      } else {
        this.setData({ passMsg: res.message || '修改失败', passSuccess: false })
      }
    } catch { this.setData({ passMsg: '修改失败', passSuccess: false }) }
    wx.hideLoading()
  },
  doLogout() {
    wx.showModal({ title: '确认退出', content: '确定要退出登录吗？', success: (r) => {
      if (r.confirm) { clearAuth(); wx.reLaunch({ url: '/pages/login/login' }) }
    }})
  },
})
