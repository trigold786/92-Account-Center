import { api } from '../../utils/api'
import { getUserId, isLoggedIn } from '../../utils/storage'
import type { Subscription, UserTier } from '../../models/types'

Page({
  data: {
    subscription: null as Subscription | null,
    tier: null as UserTier | null,
    loading: true,
  },
  onShow() {
    if (!isLoggedIn()) { wx.showToast({ title: '请先登录', icon: 'none' }); setTimeout(() => wx.navigateTo({ url: '/pages/login/login' }), 500); return }
    this.loadData()
  },
  async loadData() {
    this.setData({ loading: true })
    const uid = getUserId()
    const [subRes, tierRes] = await Promise.all([
      api.get<Subscription>(`/subscriptions/${uid}`),
      api.get<UserTier>(`/account/${uid}/tier`),
    ])
    this.setData({ subscription: subRes.data || null, tier: tierRes.data || null, loading: false })
  },
})
