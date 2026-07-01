import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AppDialogsProvider } from '@/components/ui/AppDialogs'
import type { Channel } from '@/lib/types'
import '@/lib/i18n'

// Mock at the service-module boundary (TESTING.md — no MSW).
vi.mock('@/lib/services/channels', async (importOriginal) => {
  const mod = await importOriginal<typeof import('@/lib/services/channels')>()
  return {
    ...mod,
    channelsService: {
      list: vi.fn(),
      get: vi.fn(),
      create: vi.fn(),
      test: vi.fn(),
      update: vi.fn(),
      remove: vi.fn(),
    },
  }
})

import { channelsService } from '@/lib/services/channels'
import { ChannelsSection } from './ChannelsSection'

const mockList = channelsService.list as ReturnType<typeof vi.fn>

const channel: Channel = {
  id: 1,
  type: 1,
  key: '',
  status: 1,
  name: 'worldrouter',
  weight: null,
  created_time: 0,
  test_time: 0,
  response_time: 0,
  base_url: null,
  balance: 0,
  balance_updated_time: 0,
  models: 'gpt-4o,claude-3-haiku',
  group: 'default',
  used_quota: 0,
  model_mapping: null,
  priority: 0,
  config: '',
  system_prompt: null,
}

function renderSection() {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return render(
    <QueryClientProvider client={qc}>
      <AppDialogsProvider>
        <ChannelsSection />
      </AppDialogsProvider>
    </QueryClientProvider>,
  )
}

beforeEach(() => {
  vi.clearAllMocks()
  mockList.mockResolvedValue([channel])
})

// Mutation error surfacing (test/toggle/delete + the errorText i18n fallback) is
// covered at the route level in console.channels.test.tsx; this is a co-located
// smoke test that the extracted component renders its own data.
describe('<ChannelsSection>', () => {
  it('lists channels returned by the service', async () => {
    renderSection()
    expect(await screen.findByText('worldrouter')).toBeInTheDocument()
  })

  it('renders the add-channel control', async () => {
    renderSection()
    await screen.findByText('worldrouter')
    expect(screen.getByRole('button', { name: 'Add channel' })).toBeInTheDocument()
  })
})
