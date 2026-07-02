export interface User {
  id: number
  account_id: string
  phone_number?: string
  email?: string
  tier?: string
}

export interface RFMScore {
  recency_score: number
  frequency_score: number
  monetary_score: number
  total_score: number
  segment?: string
}

export interface CreditAccount {
  balance: number
  total_earned: number
  total_consumed: number
  expires_at?: string
}

export interface Transaction {
  id: number
  amount: number
  type: 'earn' | 'consume' | 'refund'
  reason: string
  created_at: string
}

export interface Subscription {
  id: number
  user_id: number
  plan_id: string
  plan_name: string
  status: string
  current_period_start: string
  current_period_end: string
}

export interface Device {
  device_id: string
  device_name: string
  device_type: string
  is_trusted: boolean
  last_seen_at: string
}

export interface ReferralSummary {
  referral_code: string
  referral_link: string
  total_referrals: number
  total_earnings: number
}

export * from './api'
