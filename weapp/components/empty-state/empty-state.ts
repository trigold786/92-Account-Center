Component({
  options: { addGlobalClass: true },
  properties: {
    icon: { type: String, value: '' },
    title: { type: String, value: '' },
    description: { type: String, value: '' },
    actionText: { type: String, value: '' },
    actionRoute: { type: String, value: '' },
  },
  methods: {
    onAction() {
      const route = this.data.actionRoute
      if (route) {
        wx.navigateTo({ url: route })
      }
      this.triggerEvent('action')
    },
  },
})
