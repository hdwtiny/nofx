import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Loader2, X } from 'lucide-react'
import { api } from '../../lib/api'
import {
  resolveAlertSymbol,
  type MarketSymbolOption,
} from '../../lib/api/market'

const PLATFORMS = [
  'binance',
  'bybit',
  'okx',
  'bitget',
  'gate',
  'hyperliquid',
] as const

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

  const [suggestions, setSuggestions] = useState<MarketSymbolOption[]>([])
  const [searching, setSearching] = useState(false)
  const [searchError, setSearchError] = useState<string | null>(null)
  const [showSuggestions, setShowSuggestions] = useState(false)

  const [currentPrice, setCurrentPrice] = useState<number | null>(null)
  const [priceLoading, setPriceLoading] = useState(false)
  const [priceError, setPriceError] = useState<string | null>(null)

  const dropdownRef = useRef<HTMLDivElement>(null)
  const searchTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const parsedTarget = useMemo(() => Number(targetPrice), [targetPrice])
  const normalizedSymbol = useMemo(
    () => resolveAlertSymbol(symbol, platform),
    [symbol, platform]
  )

  const canSubmit =
    normalizedSymbol.length > 0 &&
    platform.trim().length > 0 &&
    Number.isFinite(parsedTarget) &&
    parsedTarget > 0

  const loadCurrentPrice = useCallback(async (sym: string, exch: string) => {
    const s = resolveAlertSymbol(sym, exch)
    const p = exch.trim().toLowerCase()
    if (!s || !p) {
      setCurrentPrice(null)
      return
    }
    setPriceLoading(true)
    setPriceError(null)
    try {
      const close = await api.getLatestClose(s, p)
      if (close == null) {
        setCurrentPrice(null)
        setPriceError('Unable to fetch current price')
      } else {
        setCurrentPrice(close)
      }
    } catch {
      setCurrentPrice(null)
      setPriceError('Unable to fetch current price')
    } finally {
      setPriceLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!normalizedSymbol || normalizedSymbol.length < 1) {
      setCurrentPrice(null)
      setPriceError(null)
      return
    }
    const timer = setTimeout(() => {
      loadCurrentPrice(normalizedSymbol, platform)
    }, 400)
    return () => clearTimeout(timer)
  }, [normalizedSymbol, platform, loadCurrentPrice])

  useEffect(() => {
    const q = symbol.trim().toUpperCase()
    if (searchTimerRef.current) clearTimeout(searchTimerRef.current)

    if (q.length < 1) {
      setSuggestions([])
      setSearchError(null)
      setSearching(false)
      return
    }

    searchTimerRef.current = setTimeout(async () => {
      setSearching(true)
      setSearchError(null)
      try {
        const rows = await api.searchSymbols(platform, q, 30)
        if (rows.length > 0) {
          setSuggestions(rows.slice(0, 20))
        } else {
          // Fallback: user may type base coin (LIT) while API lists LITUSDT
          const resolved = resolveAlertSymbol(q, platform)
          const close = await api.getLatestClose(resolved, platform)
          if (close != null) {
            setSuggestions([
              { symbol: resolved, name: resolved, price: close },
            ])
          } else {
            setSuggestions([])
            setSearchError('No matching symbols on this platform')
          }
        }
      } catch (e) {
        setSuggestions([])
        setSearchError(
          e instanceof Error ? e.message : 'Symbol search failed'
        )
      } finally {
        setSearching(false)
      }
    }, 280)

    return () => {
      if (searchTimerRef.current) clearTimeout(searchTimerRef.current)
    }
  }, [symbol, platform])

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setShowSuggestions(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  const pickSuggestion = (opt: MarketSymbolOption) => {
    setSymbol(opt.symbol)
    setShowSuggestions(false)
    if (opt.price && opt.price > 0) {
      setCurrentPrice(opt.price)
      setPriceError(null)
    }
  }

  const submit = async () => {
    if (!canSubmit) return
    setSaving(true)
    try {
      await onCreate({
        symbol: resolveAlertSymbol(symbol, platform),
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
            Platform
          </label>
          <select
            value={platform}
            onChange={(e) => {
              setPlatform(e.target.value)
              setShowSuggestions(true)
            }}
            className="w-full bg-zinc-950/80 border border-zinc-700/80 rounded-xl px-4 py-3 text-sm text-white focus:outline-none focus:border-nofx-gold/60 focus:ring-1 focus:ring-nofx-gold/30 transition-all"
          >
            {PLATFORMS.map((p) => (
              <option key={p} value={p}>
                {p}
              </option>
            ))}
          </select>
        </div>

        <div className="relative" ref={dropdownRef}>
          <label className="block text-xs font-medium text-zinc-400 mb-2">
            Symbol
          </label>
          <input
            value={symbol}
            onChange={(e) => {
              setSymbol(e.target.value.toUpperCase())
              setShowSuggestions(true)
            }}
            onFocus={() => setShowSuggestions(true)}
            placeholder="Type prefix, e.g. BTC"
            autoComplete="off"
            className="w-full bg-zinc-950/80 border border-zinc-700/80 rounded-xl px-4 py-3 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-nofx-gold/60 focus:ring-1 focus:ring-nofx-gold/30 transition-all"
          />
          {showSuggestions && symbol.trim().length > 0 && (
            <div className="absolute z-20 left-0 right-0 mt-1 max-h-48 overflow-y-auto rounded-xl border border-zinc-700/80 bg-zinc-950 shadow-xl">
              {searching ? (
                <div className="flex items-center gap-2 px-4 py-3 text-xs text-zinc-500">
                  <Loader2 size={14} className="animate-spin" />
                  Searching...
                </div>
              ) : suggestions.length === 0 ? (
                <div className="px-4 py-3 text-xs text-zinc-500">
                  {searchError || 'No matching symbols'}
                </div>
              ) : (
                suggestions.map((opt) => (
                  <button
                    key={opt.symbol}
                    type="button"
                    onClick={() => pickSuggestion(opt)}
                    className="w-full flex items-center justify-between px-4 py-2.5 text-left text-sm text-white hover:bg-zinc-800/80 transition-colors"
                  >
                    <span className="font-medium">{opt.symbol}</span>
                    {opt.price != null && opt.price > 0 && (
                      <span className="text-xs text-zinc-500 tabular-nums">
                        {opt.price.toLocaleString(undefined, {
                          maximumFractionDigits: 8,
                        })}
                      </span>
                    )}
                  </button>
                ))
              )}
            </div>
          )}
        </div>

        <div>
          <label className="block text-xs font-medium text-zinc-400 mb-2">
            Current Price
          </label>
          <div className="w-full bg-zinc-950/50 border border-zinc-800 rounded-xl px-4 py-3 text-sm text-zinc-300 tabular-nums">
            {priceLoading ? (
              <span className="inline-flex items-center gap-2 text-zinc-500">
                <Loader2 size={14} className="animate-spin" />
                Loading...
              </span>
            ) : currentPrice != null ? (
              <span className="text-white font-medium">
                {currentPrice.toLocaleString(undefined, {
                  maximumFractionDigits: 8,
                })}
              </span>
            ) : normalizedSymbol ? (
              <span className="text-zinc-500">
                {priceError || 'Enter a valid symbol'}
              </span>
            ) : (
              <span className="text-zinc-600">—</span>
            )}
          </div>
          <p className="text-[11px] text-zinc-500 mt-2">
            Latest 1m kline close (perpetual reference).
          </p>
        </div>

        <div>
          <label className="block text-xs font-medium text-zinc-400 mb-2">
            Target Price
          </label>
          <input
            value={targetPrice}
            onChange={(e) => setTargetPrice(e.target.value)}
            placeholder={
              currentPrice != null
                ? `e.g. ${(currentPrice * 1.02).toFixed(2)}`
                : 'e.g. 65000'
            }
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
