import type { PriceAlert, ServerChanConfigStatus } from '../../types'
import { API_BASE, httpClient, CryptoService } from './helpers'

export const alertsApi = {
  async listPriceAlerts(): Promise<PriceAlert[]> {
    const res = await httpClient.get<PriceAlert[]>(`${API_BASE}/price-alerts`)
    if (!res.success) throw new Error('Failed to fetch price alerts')
    return Array.isArray(res.data) ? res.data : []
  },

  async createPriceAlert(input: {
    symbol: string
    platform: string
    target_price: number
  }): Promise<{ id: string }> {
    const res = await httpClient.post<{ id: string }>(
      `${API_BASE}/price-alerts`,
      input
    )
    if (!res.success || !res.data) {
      throw new Error(res.message || 'Failed to create price alert')
    }
    return res.data
  },

  async deletePriceAlert(id: string): Promise<void> {
    const res = await httpClient.delete(`${API_BASE}/price-alerts/${id}`)
    if (!res.success) throw new Error('Failed to delete price alert')
  },

  async getServerChanStatus(): Promise<ServerChanConfigStatus> {
    const res = await httpClient.get<ServerChanConfigStatus>(
      `${API_BASE}/notifications/serverchan`
    )
    if (!res.success || !res.data) throw new Error('Failed to fetch ServerChan status')
    return res.data
  },

  async upsertServerChanConfig(input: {
    send_key: string
    enabled?: boolean
  }): Promise<void> {
    const cfg = await CryptoService.fetchCryptoConfig()
    if (!cfg.transport_encryption) {
      const res = await httpClient.put(
        `${API_BASE}/notifications/serverchan`,
        input
      )
      if (!res.success) throw new Error('Failed to save ServerChan config')
      return
    }

    const publicKey = await CryptoService.fetchPublicKey()
    await CryptoService.initialize(publicKey)

    const userId = localStorage.getItem('user_id') || ''
    const sessionId = sessionStorage.getItem('session_id') || ''

    const encryptedPayload = await CryptoService.encryptSensitiveData(
      JSON.stringify(input),
      userId,
      sessionId
    )

    const res = await httpClient.put(
      `${API_BASE}/notifications/serverchan`,
      encryptedPayload
    )
    if (!res.success) throw new Error('Failed to save ServerChan config')
  },
}

