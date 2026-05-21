export function setAriaRole(element: any, role: string) {
  if (element?.setData) {
    element.setData({ role })
  }
}

export function setAriaLabel(element: any, label: string) {
  if (element?.setData) {
    element.setData({ 'aria-label': label })
  }
}

export function setAriaLive(element: any, politeness: 'polite' | 'assertive' = 'polite') {
  if (element?.setData) {
    element.setData({ 'aria-live': politeness })
  }
}

export function setAriaHidden(element: any, hidden: boolean) {
  if (element?.setData) {
    element.setData({ 'aria-hidden': hidden ? 'true' : 'false' })
  }
}

export function focusElement(element: any) {
  if (element?.setData) {
    element.setData({ focus: true })
  }
}

export function blurElement(element: any) {
  if (element?.setData) {
    element.setData({ focus: false })
  }
}

export function manageFocus(elementId: string) {
  try {
    const query = wx.createSelectorQuery()
    query.select(`#${elementId}`).boundingClientRect()
    query.exec((res) => {
      if (res && res[0]) {
        wx.pageScrollTo({
          scrollTop: res[0].top - 50,
          duration: 300,
        })
      }
    })
  } catch {}
}

export function announce(message: string) {
  try {
    if (typeof wx !== 'undefined' && wx.createSelectorQuery) {
      const pages = getCurrentPages()
      if (pages.length > 0) {
        const page = pages[pages.length - 1] as any
        if (page.setData) {
          page.setData({
            _a11yAnnounce: message,
            _a11yTimestamp: Date.now(),
          })
        }
      }
    }
  } catch {}
}

export function createAriaAttrs(attrs: Record<string, string>): Record<string, string> {
  const result: Record<string, string> = {}
  for (const [key, value] of Object.entries(attrs)) {
    result[`aria-${key}`] = value
  }
  return result
}
