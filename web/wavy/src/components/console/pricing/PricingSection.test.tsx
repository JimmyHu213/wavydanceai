import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AppDialogsProvider } from '@/components/ui/AppDialogs'
import type { Channel } from '@/lib/types'
import '@/lib/i18n'

// Mock at the service boundary; keep the real optionsToMap so the option-list →
// ratio-map parsing PricingSection depends on is exercised, not stubbed.
vi.mock('@/lib/services/options', async (importOriginal) => {
  const mod = await importOriginal<typeof import('@/lib/services/options')>()
  return {
    ...mod,
    optionsService: { list: vi.fn(), update: vi.fn(), updateBatch: vi.fn() },
  }
})
vi.mock('@/lib/services/channels', async (importOriginal) => {
  const mod = await importOriginal<typeof import('@/lib/services/channels')>()
  return {
    ...mod,
    channelsService: { ...mod.channelsService, listAll: vi.fn() },
  }
})

import { optionsService } from '@/lib/services/options'
import { channelsService } from '@/lib/services/channels'
import { PricingSection } from './PricingSection'

const mockOptions = optionsService.list as ReturnType<typeof vi.fn>
const mockListAll = channelsService.listAll as ReturnType<typeof vi.fn>

function channel(models: string): Channel {
  return {
    id: 1, type: 1, key: '', status: 1, name: 'worldrouter', weight: null,
    created_time: 0, test_time: 0, response_time: 0, base_url: null, balance: 0,
    balance_updated_time: 0, models, group: 'default', used_quota: 0,
    model_mapping: null, priority: 0, config: '', system_prompt: null,
  }
}

function renderSection() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return render(
    <QueryClientProvider client={qc}>
      <AppDialogsProvider>
        <PricingSection />
      </AppDialogsProvider>
    </QueryClientProvider>,
  )
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe('<PricingSection>', () => {
  it('renders the editor scoped to channel-provided models, flagging unpriced ones', async () => {
    mockOptions.mockResolvedValue([
      { key: 'GroupRatio', value: '{"default":1}' },
      { key: 'ModelRatio', value: '{"gpt-4o":1.25}' },
      { key: 'CompletionRatio', value: '{"gpt-4o":4}' },
    ])
    // Channels provide two models; only gpt-4o is priced → new-model is unpriced.
    mockListAll.mockResolvedValue([channel('gpt-4o,new-model')])

    renderSection()

    // The priced provided model renders its derived row…
    expect(await screen.findByLabelText('gpt-4o input price')).toHaveValue('2.5')
    // …and the provided-but-unpriced model is surfaced.
    expect(await screen.findByText(/1 unpriced/i)).toBeInTheDocument()
  })

  it('shows a fetch error when the options request fails', async () => {
    mockOptions.mockRejectedValue(new Error('boom'))
    mockListAll.mockResolvedValue([channel('gpt-4o')])

    renderSection()

    expect(await screen.findByText(/failed to load pricing options/i)).toBeInTheDocument()
  })
})
