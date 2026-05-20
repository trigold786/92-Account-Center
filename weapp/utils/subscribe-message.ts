interface TemplateMessage {
  templateType: string
  data: Record<string, string>
}

const TEMPLATE_IDS: Record<string, string> = {
  subscription_expiring: 'TEMPLATE_ID_1',
  payment_success: 'TEMPLATE_ID_2',
  referral_bonus: 'TEMPLATE_ID_3',
  tier_upgrade: 'TEMPLATE_ID_4',
}

export function requestSubscribeMessage(templateTypes: string[]): Promise<boolean> {
  return new Promise((resolve, reject) => {
    const tmplIds = templateTypes.map(t => TEMPLATE_IDS[t]).filter(Boolean)
    if (tmplIds.length === 0) {
      reject(new Error('No valid template IDs'))
      return
    }
    wx.requestSubscribeMessage({
      tmplIds,
      success(res) {
        const allAccepted = templateTypes.every(t => {
          const id = TEMPLATE_IDS[t]
          return res[id] === 'accept'
        })
        resolve(allAccepted)
      },
      fail(err) {
        reject(err)
      },
    })
  })
}

export function sendTemplateMessage(templateType: string, data: Record<string, string>) {
  return new Promise<void>((resolve, reject) => {
    wx.cloud.callFunction({
      name: 'sendTemplateMessage',
      data: { templateType, data } as TemplateMessage,
      success() {
        resolve()
      },
      fail(err) {
        reject(err)
      },
    })
  })
}
