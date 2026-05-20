import { useMemo, useState } from 'react'
import { X } from 'lucide-react'

export function ServerChanConfigModal(props: {
  onClose: () => void
  onSave: (input: { send_key: string; enabled?: boolean }) => Promise<void>
}) {
  const { onClose, onSave } = props
  const [sendKey, setSendKey] = useState('')
  const [enabled, setEnabled] = useState(true)
  const [saving, setSaving] = useState(false)

  const canSubmit = useMemo(() => sendKey.trim().length > 0, [sendKey])

  const submit = async () => {
    if (!canSubmit) return
    setSaving(true)
    try {
      await onSave({ send_key: sendKey.trim(), enabled })
      onClose()
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="w-full max-w-lg bg-zinc-900 border border-zinc-800 rounded-2xl overflow-hidden">
      <div className="flex items-center justify-between px-5 py-4 border-b border-zinc-800">
        <div>
          <h3 className="text-sm font-semibold text-white">ServerChan</h3>
          <p className="text-xs text-zinc-500 mt-0.5">
            Configure SendKey for alerts
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
            SendKey
          </label>
          <input
            value={sendKey}
            onChange={(e) => setSendKey(e.target.value)}
            placeholder="e.g. SCTxxxxxxxxxxxxxxxx"
            className="w-full bg-zinc-950/80 border border-zinc-700/80 rounded-xl px-4 py-3 text-sm text-white placeholder-zinc-600 focus:outline-none focus:border-nofx-gold/60 focus:ring-1 focus:ring-nofx-gold/30 transition-all"
          />
          <p className="text-[11px] text-zinc-500 mt-2">
            We never display your SendKey after saving.
          </p>
        </div>

        <label className="flex items-center gap-2 text-sm text-zinc-300">
          <input
            type="checkbox"
            checked={enabled}
            onChange={(e) => setEnabled(e.target.checked)}
          />
          Enable notifications
        </label>
      </div>

      <div className="px-5 pb-5">
        <button
          onClick={submit}
          disabled={!canSubmit || saving}
          className="w-full bg-nofx-gold hover:bg-yellow-400 active:scale-[0.98] text-black font-semibold py-3 rounded-xl text-sm transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {saving ? 'Saving...' : 'Save'}
        </button>
      </div>
    </div>
  )
}

