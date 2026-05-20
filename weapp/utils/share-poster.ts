type PosterTemplate = 'tech' | 'minimal' | 'festive'

interface PosterConfig {
  template: PosterTemplate
  userName: string
  referralCode: string
  qrCodeUrl: string
  avatarUrl?: string
}

interface TemplateStyle {
  background: string
  titleColor: string
  subtitleColor: string
  accentGradient: [string, string]
}

const TEMPLATE_STYLES: Record<PosterTemplate, TemplateStyle> = {
  tech: {
    background: '#0D1117',
    titleColor: '#E6EDF3',
    subtitleColor: '#8B949E',
    accentGradient: ['#6C63FF', '#00D4FF'],
  },
  minimal: {
    background: '#FFFFFF',
    titleColor: '#1A1A2E',
    subtitleColor: '#666666',
    accentGradient: ['#6C63FF', '#8B5CF6'],
  },
  festive: {
    background: '#1A0A2E',
    titleColor: '#FFD700',
    subtitleColor: '#E6EDF3',
    accentGradient: ['#FF6B6B', '#FFD93D'],
  },
}

const TEMPLATE_PREVIEWS: Record<PosterTemplate, string> = {
  tech: '/assets/poster-preview-tech.png',
  minimal: '/assets/poster-preview-minimal.png',
  festive: '/assets/poster-preview-festive.png',
}

function getTemplatePreview(template: PosterTemplate): string {
  return TEMPLATE_PREVIEWS[template]
}

function getAllTemplatePreviews(): { template: PosterTemplate; preview: string }[] {
  return (Object.keys(TEMPLATE_STYLES) as PosterTemplate[]).map((t) => ({
    template: t,
    preview: TEMPLATE_PREVIEWS[t],
  }))
}

function generatePoster(config: PosterConfig): Promise<string> {
  return new Promise((resolve, reject) => {
    const canvas = wx.createOffscreenCanvas({
      type: '2d',
      width: 750,
      height: 1334,
    })
    const ctx = canvas.getContext('2d')
    const style = TEMPLATE_STYLES[config.template]

    ctx.fillStyle = style.background
    ctx.fillRect(0, 0, 750, 1334)

    const gradient = ctx.createLinearGradient(0, 0, 750, 1334)
    gradient.addColorStop(0, style.accentGradient[0])
    gradient.addColorStop(1, style.accentGradient[1])
    ctx.fillStyle = gradient
    ctx.fillRect(0, 0, 750, 4)

    ctx.fillStyle = style.titleColor
    ctx.font = 'bold 40px sans-serif'
    ctx.textAlign = 'center'
    ctx.fillText('Neuro AI', 375, 200)

    ctx.fillStyle = style.subtitleColor
    ctx.font = '24px sans-serif'
    ctx.fillText('Join me on Neuro AI', 375, 260)

    ctx.fillStyle = style.titleColor
    ctx.font = 'bold 32px sans-serif'
    ctx.fillText(config.userName, 375, 340)

    ctx.fillStyle = style.subtitleColor
    ctx.font = '20px sans-serif'
    ctx.fillText('Invitation Code', 375, 400)

    ctx.fillStyle = style.accentGradient[0]
    ctx.font = 'bold 48px sans-serif'
    ctx.fillText(config.referralCode, 375, 470)

    const qrImage = canvas.createImage()
    qrImage.src = config.qrCodeUrl
    qrImage.onload = () => {
      ctx.drawImage(qrImage, 275, 540, 200, 200)

      ctx.fillStyle = style.subtitleColor
      ctx.font = '18px sans-serif'
      ctx.fillText('Scan to join', 375, 790)

      wx.canvasToTempFilePath({
        canvas,
        success: (res) => resolve(res.tempFilePath),
        fail: (err) => reject(err),
      })
    }
    qrImage.onerror = () => {
      wx.canvasToTempFilePath({
        canvas,
        success: (res) => resolve(res.tempFilePath),
        fail: (err) => reject(err),
      })
    }
  })
}

function savePosterToAlbum(filePath: string): Promise<void> {
  return new Promise((resolve, reject) => {
    wx.saveImageToPhotosAlbum({
      filePath,
      success: () => resolve(),
      fail: (err) => reject(err),
    })
  })
}

export {
  generatePoster,
  savePosterToAlbum,
  getTemplatePreview,
  getAllTemplatePreviews,
  TEMPLATE_STYLES,
}
export type { PosterTemplate, PosterConfig, TemplateStyle }
