import { useTranslation } from 'react-i18next'
import { VendorIcon } from '@/components/landing/VendorIcons'

const TOP = [
  { name: 'Claude Opus 4.6', icon: 'ic-anthropic', tok: 184.2, share: 26 },
  { name: 'GPT-5.2', icon: 'ic-openai', tok: 171.8, share: 24 },
  { name: 'DeepSeek-V4', icon: 'ic-deepseek', tok: 96.4, share: 14 },
  { name: 'Gemini 3 Pro', icon: 'ic-google', tok: 88.1, share: 12 },
  { name: 'Qwen3-Max', icon: 'ic-qwen', tok: 64.0, share: 9 },
]

export function TopModelsPanel() {
  const { t } = useTranslation()
  const max = Math.max(...TOP.map((m) => m.tok))

  return (
    <div className="flex h-full flex-col rounded-2xl border border-[color:var(--border)] bg-[color:var(--surface)] p-5">
      <header className="mb-5 flex items-baseline justify-between">
        <h3 className="font-display text-base font-bold tracking-[-0.3px]">{t('console.dash.topModels')}</h3>
        <span className="font-mono text-xs tracking-[1.5px] text-[color:var(--muted)] uppercase">24h</span>
      </header>

      <div className="flex flex-1 flex-col gap-3.5">
        {TOP.map((m, i) => (
          <div key={m.name} className="grid grid-cols-[20px_24px_1fr_auto] items-center gap-2.5">
            <span className="font-mono text-xs font-bold text-[color:var(--muted)]/70">{String(i + 1).padStart(2, '0')}</span>
            <VendorIcon id={m.icon} size={20} />
            <div className="min-w-0">
              <div className="overflow-hidden text-ellipsis whitespace-nowrap text-sm font-medium">{m.name}</div>
              <div className="mt-1 h-1 overflow-hidden rounded-full bg-[color:var(--border)]/60">
                <span
                  className="block h-full rounded-full bg-gradient-to-r from-[#3FB3D9] via-[#4ED4DC] to-[#B5ECF2]"
                  style={{ width: `${(m.tok / max) * 100}%`, transition: 'width 700ms cubic-bezier(.22,.8,.3,1)' }}
                />
              </div>
            </div>
            <span className="text-right font-mono text-xs tabular-nums text-[color:var(--muted)]">
              {m.tok.toFixed(1)}B
            </span>
          </div>
        ))}
      </div>

      <footer className="mt-5 border-t border-[color:var(--border)] pt-3 text-center">
        <a href="#" className="font-mono text-xs uppercase tracking-[1.5px] text-[color:var(--cyan)] hover:underline">
          {t('console.dash.viewAll')} →
        </a>
      </footer>
    </div>
  )
}
