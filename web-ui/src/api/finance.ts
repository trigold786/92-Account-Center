import client from './client'

export interface PageQuery {
  page?: number
  page_size?: number
  status?: string
  user_id?: number | string
  payment_method?: string
  start_time?: string
  end_time?: string
}

export interface CreateInvoicePayload {
  user_id: number
  order_id: number
  title: string
  tax_id?: string
  email: string
  amount: number
}

export interface Refund {
  id: number
  order_id: number
  user_id: number
  amount: number
  reason?: string
  status: string
  refund_no?: string
  provider?: string
  provider_refund_id?: string
  provider_status?: string
  provider_error?: string
  approved_at?: string
  failed_at?: string
}

export function listOrders(params: PageQuery = {}) {
  return client.get('/orders', { params })
}

export function listRefunds(params: PageQuery = {}) {
  return client.get('/refunds', { params })
}

export function approveRefund(id: number) {
  return client.put(`/refunds/${id}/approve`)
}

export function rejectRefund(id: number, note: string) {
  return client.put(`/refunds/${id}/reject`, { note })
}

export function listInvoices(params: PageQuery = {}) {
  return client.get('/invoices', { params })
}

export function createInvoice(payload: CreateInvoicePayload) {
  return client.post('/invoices', payload)
}

export function reconcile(provider: string, date: string) {
  return client.post('/payment/reconcile', { provider, date })
}

export function getReconciliationReport(reportId: string) {
  return client.get(`/payment/reconcile/${reportId}`)
}
