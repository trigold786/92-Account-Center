Component({
  options: { addGlobalClass: true },
  properties: { text: { type: String, value: '' }, disabled: { type: Boolean, value: false }, loading: { type: Boolean, value: false } },
  methods: { onClick() { if (!this.data.disabled) this.triggerEvent('tap') } }
})
