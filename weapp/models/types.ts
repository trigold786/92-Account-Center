export interface RFMScore {
  recency_score: number
  frequency_score: number
  monetary_score: number
  total_score: number
  segment?: string
  last_calculated_at?: string
}

export interface CreditAccount {
  balance: number
  total_earned: number
  total_consumed: number
  expires_at?: string
}

export interface CreditTransaction {
  id: number
  amount: number
  type: 'earn' | 'consume' | 'refund'
  reason: string
  created_at: string
}

export interface ReferralSummary {
  referral_code: string
  referral_link: string
  total_referrals: number
  total_earnings: number
  level1_count: number
  level2_count: number
}

export interface Subscription {
  id: number
  user_id: number
  plan_id: string
  plan_name: string
  status: 'active' | 'expired' | 'cancelled' | 'trialing'
  current_period_start: string
  current_period_end: string
  canceled_at?: string
  trial_end?: string
}

export interface UserTier {
  tier: string
  benefits: string[]
}

export interface DeviceInfo {
  device_id: string
  device_name: string
  device_type: string
  is_trusted: boolean
  last_seen_at: string
  created_at: string
}

export interface RiskEvent {
  id: number
  event_type: string
  risk_score: number
  risk_level: string
  detail: string
  created_at: string
}

export interface User {
  id: number
  account_id: string
  phone_number?: string
  email?: string
  tier?: string
}

export interface ApiResponse<T = any> {
  code: number
  message: string
  data?: T
  total?: number
}
