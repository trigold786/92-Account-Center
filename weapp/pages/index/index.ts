import { api } from '../../utils/api'
import { isLoggedIn, getUserId } from '../../utils/storage'
import type { RFMScore, User } from '../../models/types'

Page({
  data: {
    loggedIn: false,
    user: null as User | null,
    rfm: null as RFMScore | null,
    rfmSegment: '',
    loading: true,
  },
  onShow() {
    const loggedIn = isLoggedIn()
    this.setData({ loggedIn })
    if (loggedIn) this.loadData()
    else this.setData({ loading: false })
  },
  async loadData() {
    this.setData({ loading: true })
    const userId = getUserId()
    const [userRes, rfmRes] = await Promise.all([
      api.get<User>(`/account/${userId}/tier`),
      api.get<RFMScore>(`/data/rfm/${userId}`),
    ])
    this.setData({
      user: userRes.data || null,
      rfm: rfmRes.data || null,
      loading: false,
    })
    if (rfmRes.data) {
      const score = rfmRes.data.total_score
      let segment = ''
      if (score >= 12) segment = '⭐⭐ 高价值用户'
      else if (score >= 8) segment = '⭐ 成长用户'
      else if (score >= 4) segment = '待激活用户'
      else segment = '新用户'
      this.setData({ rfmSegment: segment })
    }
  },
  goLogin() { wx.navigateTo({ url: '/pages/login/login' }) },
  goCredits() { wx.switchTab({ url: '/pages/credits/credits' }) },
  goSubscription() { wx.switchTab({ url: '/pages/subscription/subscription' }) },
})
