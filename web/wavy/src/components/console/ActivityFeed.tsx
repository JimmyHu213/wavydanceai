import { useTranslation } from 'react-i18next'
import { KeyRound, AlertTriangle, ArrowUpRight, UserPlus, Boxes, type LucideIcon } from 'lucide-react'
import { cn } from '@/lib/cn'

type Tone = 'info' | 'success' | 'warn' | 'coral'

type Item = {
  icon: LucideIcon
  tone: Tone
  titleKey: string
  metaKey: string
  ts: string
}

const ITEMS: Item[] = [
  { icon: KeyRound, tone: 'success', titleKey: 'console.dash.act.key', metaKey: 'console.dash.act.keyMeta', ts: '2m' },
  { icon: AlertTriangle, tone: 'coral', titleKey: 'console.dash.act.fail', metaKey: 'console.dash.act.failMeta', ts: '6m' },
  { icon: ArrowUpRight, tone: 'info', titleKey: 'console.dash.act.quota', metaKey: 'console.dash.act.quotaMeta', ts: '18m' },
  { icon: UserPlus, tone: 'success', titleKey: 'console.dash.act.upgrade', metaKey: 'console.dash.act.upgradeMeta', ts: '42m' },
  { icon: Boxes, tone: 'info', titleKey: 'console.dash.act.routing', metaKey: 'console.dash.act.routingMeta', ts: '1h' },
]

const TONE: Record<Tone, { bg: string; text: string }> = {
  info: { bg: 'var(--cyan)', text: 'var(--cyan)' },
  success: { bg: 'var(--live)', text: 'var(--live)' },
  warn: { bg: '#F5C26B', text: '#F5C26B' },
  coral: { bg: 'var(--coral)', text: 'var(--coral)' },
}

export function ActivityFeed() {
  const { t } = useTranslation()
  return (
    <div className="rounded-2xl border border-[color:var(--border)] bg-[color:var(--surface)] p-5">
      <header className="mb-5 flex items-baseline justify-between">
        <h3 className="font-display text-base font-bold tracking-[-0.3px]">{t('console.dash.activity')}</h3>
        <a href="#" className="font-mono text-xs uppercase tracking-[1.5px] text-[color:var(--cyan)] hover:underline">
          {t('console.dash.openLog')} →
        </a>
      </header>

      <ol className="relative">
        <span className="absolute bottom-3 left-[15px] top-3 w-px bg-[color:var(--border)]" />
        {ITEMS.map((item, i) => {
          const Icon = item.icon
          const c = TONE[item.tone]
          return (
            <li key={i} className="relative grid grid-cols-[32px_1fr_auto] items-start gap-3 py-2.5">
              <span
                className={cn(
                  'relative z-10 flex h-8 w-8 items-center justify-center rounded-full border bg-[color:var(--surface)]',
                )}
                style={{ borderColor: `color-mix(in srgb, ${c.bg} 50%, transparent)` }}
              >
                <Icon className="h-3.5 w-3.5" style={{ color: c.text }} strokeWidth={2.25} />
              </span>
              <div className="min-w-0">
                <div className="text-sm font-medium text-[color:var(--text)]">{t(item.titleKey)}</div>
                <div className="mt-0.5 font-mono text-xs text-[color:var(--muted)]">{t(item.metaKey)}</div>
              </div>
              <span className="font-mono text-xs tabular-nums text-[color:var(--muted)]/70">{item.ts} ago</span>
            </li>
          )
        })}
      </ol>
    </div>
  )
}
