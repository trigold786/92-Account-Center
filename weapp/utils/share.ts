interface ShareOptions {
  title: string
  path: string
  imageUrl?: string
  inviterId?: string
}

export function getSharePath(path: string, inviterId?: string): string {
  if (!inviterId) return path
  const separator = path.includes('?') ? '&' : '?'
  return `${path}${separator}inviter_id=${inviterId}`
}

export function onShareAppMessage(options: ShareOptions) {
  const inviterId = wx.getStorageSync('inviter_id') || ''
  return {
    title: options.title,
    path: getSharePath(options.path, inviterId || options.inviterId),
    imageUrl: options.imageUrl || '',
  }
}

export function onShareTimeline(options: ShareOptions) {
  const inviterId = wx.getStorageSync('inviter_id') || ''
  return {
    title: options.title,
    query: inviterId ? `inviter_id=${inviterId || options.inviterId}` : '',
    imageUrl: options.imageUrl || '',
  }
}
