import { httpClient } from './helpers'

export type MarketSymbolOption = {
  symbol: string
  name: string
  price?: number
}

export const marketApi = {
  async searchSymbols(
    exchange: string,
    q: string,
    limit = 20
  ): Promise<MarketSymbolOption[]> {
    const params = new URLSearchParams({
      exchange: exchange.trim(),
      q: q.trim(),
      limit: String(limit),
    })
    const res = await httpClient.request<{ symbols: MarketSymbolOption[] }>(
      `/api/market/symbols/search?${params.toString()}`,
      { silent: true }
    )
    if (!res.success) throw new Error('Failed to search symbols')
    const body = res.data as { symbols?: MarketSymbolOption[] } | undefined
    return Array.isArray(body?.symbols) ? body.symbols : []
  },

  async getLatestClose(
    symbol: string,
    exchange: string
  ): Promise<number | null> {
    const params = new URLSearchParams({
      symbol: symbol.trim(),
      exchange: exchange.trim(),
      interval: '1m',
      limit: '1',
    })
    const res = await httpClient.request<
      Array<{ close: number; openTime: number }>
    >(`/api/klines?${params.toString()}`, { silent: true })
    if (!res.success || !Array.isArray(res.data) || res.data.length === 0) {
      return null
    }
    const last = res.data[res.data.length - 1]
    return typeof last.close === 'number' && last.close > 0 ? last.close : null
  },
}
