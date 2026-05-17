App<IMPOpts>({
  globalData: {
    userInfo: null as any | null,
    token: '',
    refreshToken: '',
    userId: 0,
    accountId: '',
    baseURL: 'https://api.accountcenter.com/api/v1',
  },
  onLaunch() {
    const token = wx.getStorageSync('access_token')
    const refreshToken = wx.getStorageSync('refresh_token')
    if (token) {
      this.globalData.token = token
      this.globalData.refreshToken = refreshToken
    }
  },
})
