import client from './client'

export interface TierFeature {
  name: string
  description: string
  included: boolean
}

export interface TierEntitlement {
  feature_code: string
  quota: number
  unit: string
}

export interface PricingTier {
  tier_level: number
  name: string
  monthly_price: number
  yearly_price: number
  features: TierFeature[]
  entitlements: TierEntitlement[]
}

export interface PricingPageData {
  tiers: PricingTier[]
  credit_rate: number
  max_credit_discount: number
}

export interface CreditDiscountResult {
  original_price: number
  credits_used: number
  credit_value: number
  final_price: number
  discount_percent: number
}

export function getPricing() {
  return client.get('/pricing')
}

export function calculateDiscount(price: number, creditsUsed: number) {
  return client.post('/pricing/calculate-discount', { price, credits_used: creditsUsed })
}
