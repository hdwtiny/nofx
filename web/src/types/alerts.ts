export type PriceAlertStatus = 'pending' | 'triggered' | 'cancelled'
export type PriceAlertDirection = 'up' | 'down'

export type PriceAlert = {
  id: string
  symbol: string
  platform: string
  target_price: number
  reference_price: number
  direction: PriceAlertDirection
  status: PriceAlertStatus
  triggered_at?: string | null
  triggered_price?: number | null
  created_at: string
}

export type ServerChanConfigStatus = {
  enabled: boolean
  configured: boolean
}

