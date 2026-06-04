import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { ArrowUpRight, Plus } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { StatCard } from '@/components/console/StatCard'
import { AreaChart } from '@/components/console/AreaChart'
import { TopModelsPanel } from '@/components/console/TopModelsPanel'
import { ActivityFeed } from '@/components/console/ActivityFeed'

export const Route = createFileRoute('/console/')({
  component: Dashboard,
})

// 7 days of synthetic data
const REQUEST_VOLUME = [
  { label: 'MON', value: 842_000 },
  { label: 'TUE', value: 911_000 },
  { label: 'WED', value: 1_024_000 },
  { label: 'THU', value: 968_000 },
  { label: 'FRI', value: 1_180_000 },
  { label: 'SAT', value: 1_310_000 },
  { label: 'SUN', value: 1_242_000 },
]

export default function Dashboard() {
  const { t } = useTranslation()

  return (
    <div className="mx-auto w-full max-w-[1400px] flex-1 px-6 py-8 lg:px-10">
      <style>{`
        @keyframes wavy-rise {
          from { opacity: 0; transform: translateY(14px); }
          to   { opacity: 1; transform: translateY(0); }
        }
        .wavy-rise { animation: wavy-rise .55s cubic-bezier(.22,.8,.3,1) both; }
      `}</style>

      {/* Page header */}
      <header className="wavy-rise mb-8 flex flex-wrap items-end justify-between gap-4" style={{ animationDelay: '0ms' }}>
        <div>
          <div className="font-mono text-xs uppercase tracking-[2.5px] text-[color:var(--cyan)]">
            {t('console.dash.kicker')}
          </div>
          <h1 className="mt-2 font-display text-[2rem] font-bold leading-tight tracking-[-1px]">
            {t('console.dash.title')}
          </h1>
          <p className="mt-1.5 text-sm text-[color:var(--muted)]">{t('console.dash.lead')}</p>
        </div>
        <div className="flex gap-2.5">
          <Button variant="ghost" size="sm">
            <ArrowUpRight className="h-3.5 w-3.5" />
            {t('console.dash.export')}
          </Button>
          <Button size="sm">
            <Plus className="h-3.5 w-3.5" />
            {t('console.dash.newKey')}
          </Button>
        </div>
      </header>

      {/* KPI grid */}
      <section className="mb-7 grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <div className="wavy-rise" style={{ animationDelay: '40ms' }}>
          <StatCard
            kicker={t('console.dash.kpi.requests')}
            value="1.24"
            unit="M"
            delta={8.2}
            spark={[8.4, 9.1, 10.2, 9.7, 11.8, 13.1, 12.4]}
          />
        </div>
        <div className="wavy-rise" style={{ animationDelay: '90ms' }}>
          <StatCard
            kicker={t('console.dash.kpi.tokens')}
            value="38.7"
            unit="B"
            delta={12.0}
            spark={[24, 28, 30, 29, 33, 36, 38.7]}
          />
        </div>
        <div className="wavy-rise" style={{ animationDelay: '140ms' }}>
          <StatCard
            kicker={t('console.dash.kpi.activeModels')}
            value="187"
            delta={1.6}
            spark={[170, 174, 179, 178, 182, 185, 187]}
          />
        </div>
        <div className="wavy-rise" style={{ animationDelay: '190ms' }}>
          <StatCard
            kicker={t('console.dash.kpi.spend')}
            value="$4,328"
            delta={-2.1}
            spark={[4500, 4480, 4420, 4350, 4310, 4340, 4328]}
          />
        </div>
      </section>

      {/* Chart + Top models */}
      <section className="mb-7 grid grid-cols-1 gap-4 xl:grid-cols-12">
        <div
          className="wavy-rise xl:col-span-8 rounded-2xl border border-[color:var(--border)] bg-[color:var(--surface)] p-5"
          style={{ animationDelay: '240ms' }}
        >
          <header className="mb-1 flex items-baseline justify-between">
            <div>
              <h3 className="font-display text-base font-bold tracking-[-0.3px]">{t('console.dash.chart.title')}</h3>
              <p className="mt-1 text-xs text-[color:var(--muted)]">{t('console.dash.chart.sub')}</p>
            </div>
            <div className="flex gap-1.5">
              {['24H', '7D', '30D', 'ALL'].map((r, i) => (
                <button
                  key={r}
                  type="button"
                  className={
                    i === 1
                      ? 'rounded-md bg-gradient-to-r from-[#3FB3D9] to-[#4ED4DC] px-2.5 py-1 font-mono text-xs font-bold tracking-[1px] text-[#052832]'
                      : 'rounded-md border border-[color:var(--border)] px-2.5 py-1 font-mono text-xs tracking-[1px] text-[color:var(--muted)] transition hover:border-[color:var(--cyan)] hover:text-[color:var(--text)]'
                  }
                >
                  {r}
                </button>
              ))}
            </div>
          </header>
          <div className="mt-4">
            <AreaChart data={REQUEST_VOLUME} />
          </div>
        </div>

        <div className="wavy-rise xl:col-span-4" style={{ animationDelay: '290ms' }}>
          <TopModelsPanel />
        </div>
      </section>

      {/* Activity feed full width */}
      <section className="wavy-rise" style={{ animationDelay: '340ms' }}>
        <ActivityFeed />
      </section>
    </div>
  )
}
