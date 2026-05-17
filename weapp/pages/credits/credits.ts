import { api } from '../../utils/api'
import { getUserId, isLoggedIn } from '../../utils/storage'
import type { CreditAccount, CreditTransaction, ReferralSummary } from '../../models/types'

Page({
  data: {
    loggedIn: false,
    account: null as CreditAccount | null,
    transactions: [] as CreditTransaction[],
    referral: null as ReferralSummary | null,
    loading: true,
    earning: false,
  },
  onShow() {
    if (!isLoggedIn()) { wx.showToast({ title: '请先登录', icon: 'none' }); setTimeout(() => wx.navigateTo({ url: '/pages/login/login' }), 500); return }
    this.loadData()
  },
  async loadData() {
    this.setData({ loading: true })
    const uid = getUserId()
    const [acctRes, txnRes, refRes] = await Promise.all([
      api.get<CreditAccount>(`/credits/${uid}/account`),
      api.get<CreditTransaction[]>(`/credits/${uid}/transactions`),
      api.get<ReferralSummary>(`/referral/${uid}/summary`),
    ])
    this.setData({
      loggedIn: true,
      account: acctRes.data || null,
      transactions: (txnRes.data || []).slice(0, 20),
      referral: refRes.data || null,
      loading: false,
    })
  },
  async doEarn() {
    if (this.data.earning) return
    this.setData({ earning: true })
    try {
      const res = await api.post('/credits/earn', { user_id: getUserId(), amount: 1, reason: 'daily_checkin' })
      if (res.code === 0) {
        wx.showToast({ title: '签到成功 +1', icon: 'success' })
        this.loadData()
      } else {
        wx.showToast({ title: res.message || '签到失败', icon: 'none' })
      }
    } catch { wx.showToast({ title: '签到失败', icon: 'none' }) }
    this.setData({ earning: false })
  },
  copyReferralLink() {
    if (this.data.referral?.referral_link) {
      wx.setClipboardData({ data: this.data.referral.referral_link })
    }
  },
})
