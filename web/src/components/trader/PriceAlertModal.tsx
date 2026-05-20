import { useMemo, useState } from 'react'
import { X } from 'lucide-react'

export function PriceAlertModal(props: {
  onClose: () => void
  onCreate: (input: {
    symbol: string
    platform: string
    target_price: number
  }) => Promise<void>
  defaultPlatform?: string
  language?: string
}) {
  const { onClose, onCreate, defaultPlatform } = props
  const [symbol, setSymbol] = useState('')
  const [platform, setPlatform] = useState(defaultPlatform || 'binance')
  const [targetPrice, setTargetPrice] = useState('')
  const [saving, setSaving] = useState(false)

  const parsedTarget = useMemo(() => Number(targetPrice), [targetPrice])
  const canSubmit =
    symbol.trim().length > 0 &&
    platform.trim().length > 0 &&
    Number.isFinite(parsedTarget) &&
    parsedTarget > 0

  const submit = async () => {
    if (!canSubmit) return
    setSaving(true)
    try {
      await onCreate({
        symbol: symbol.trim(),
        platform: platform.trim(),
        target_price: parsedTarget,
      })
      onClose()
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="w-full max-w-lg bg-zinc-900 border border-zinc-800 rounded-2xl overflow-hidden">
      <div className="flex items-center justify-between px-5 py-4 border-b border-zinc-800">
        <div>
          <h3 className="text-sm font-semibold text-white">Create Price Alert</h3>
          <p className="text-xs text-zinc-500 mt-0.5">
            One-time alert via ServerChan
          </p>
        </div>
        <button
          onClick={onClose}
          className="text-zinc-500 hover:text-zinc-300 transition-colors"
        >
          <X size={16} />
        </button>
      </div>

      <div className="p-5 space-y-4">
        <div>
          <label className="block text-xs font-medium text-zinc-400 mb-2">
            Symbol
          </label>
          <input
            value={symbol}
            onChange={(e) => setSymbol(e.target.value)}
            placeholder="e.g. BTCUSDT"
            className="w-full bg-zinc-950/80 border border-zinc-700/80 rounded-xl px-4 py-3 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-nofx-gold/60 focus:ring-1 focus:ring-nofx-gold/30 transition-all"
          />
        </div>

        <div>
          <label className="block text-xs font-medium text-zinc-400 mb-2">
            Platform
          </label>
          <input
            value={platform}
            onChange={(e) => setPlatform(e.target.value)}
            placeholder="e.g. binance / bybit / okx / hyperliquid"
            className="w-full bg-zinc-950/80 border border-zinc-700/80 rounded-xl px-4 py-3 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-nofx-gold/60 focus:ring-1 focus:ring-nofx-gold/30 transition-all"
          />
        </div>

        <div>
          <label className="block text-xs font-medium text-zinc-400 mb-2">
            Target Price
          </label>
          <input
            value={targetPrice}
            onChange={(e) => setTargetPrice(e.target.value)}
            placeholder="e.g. 65000"
            inputMode="decimal"
            className="w-full bg-zinc-950/80 border border-zinc-700/80 rounded-xl px-4 py-3 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-nofx-gold/60 focus:ring-1 focus:ring-nofx-gold/30 transition-all"
          />
          <p className="text-[11px] text-zinc-500 mt-2">
            Trigger uses latest 1m kline high/low range (limit=2).
          </p>
        </div>
      </div>

      <div className="px-5 pb-5">
        <button
          onClick={submit}
          disabled={!canSubmit || saving}
          className="w-full bg-nofx-gold hover:bg-yellow-400 active:scale-[0.98] text-black font-semibold py-3 rounded-xl text-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {saving ? 'Creating...' : 'Create Alert'}
        </button>
      </div>
    </div>
  )
}

