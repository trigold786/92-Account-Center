export interface ApiResponse<T> {
  data: T
  message?: string
  error?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  page_size: number
}

export interface AdminUser {
  id: number
  user_id: number
  account_id: string
  phone_number?: string
  email?: string
  tier: string
  identity_tier?: number
  status: string
  mfa_enabled: boolean
  created_at: string
  updated_at: string
}

export interface Role {
  id: number
  name: string
  description: string
  is_system: boolean
  created_at: string
}

export interface Permission {
  id: number
  permission: string
  resource: string
  action: string
  description: string
}

export interface PaymentOrder {
  id: number
  order_no: string
  user_id: number
  amount: number
  currency: string
  status: string
  payment_method?: string
  created_at: string
}

export interface Refund {
  id: number
  refund_no: string
  order_id: number
  amount: number
  status: string
  reason?: string
  created_at: string
}

export interface Subscription {
  id: number
  user_id: number
  plan: string
  status: string
  start_date: string
  end_date?: string
}

export interface PricingPlan {
  id: string
  name: string
  monthly_price: number
  yearly_price: number
  features: string[]
}

export interface Device {
  id: number
  user_id: number
  device_id: string
  device_token: string
  platform: string
  device_name?: string
  device_type?: string
  is_trusted?: boolean
  last_active_at: string
  last_seen_at?: string
}

export interface Transaction {
  id: number
  user_id: number
  type: string
  amount: number
  description?: string
  created_at: string
}

export interface KYCStatus {
  id: number
  user_id: number
  status: string
  submitted_at: string
  reviewed_at?: string
}
