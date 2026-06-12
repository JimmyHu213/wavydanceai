import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { StatusInfo } from '@/lib/services/status'
import '@/lib/i18n'

// Route-level component: stub the router surface so we can render the page
// without a full <RouterProvider> (no router test helper yet — TESTING.md).
vi.mock('@tanstack/react-router', () => ({
  createFileRoute: () => (options: unknown) => ({ options }),
  Link: ({ children }: { children?: React.ReactNode }) => <a>{children}</a>,
  redirect: vi.fn(),
  useNavigate: () => vi.fn(),
  useSearch: () => ({ next: '/console' }),
}))

vi.mock('@/lib/services/status', () => ({
  statusService: { get: vi.fn() },
}))

vi.mock('@/components/passkey/passkey-ceremonies', () => ({
  isWebAuthnSupported: vi.fn(() => true),
  beginPasskeyRegistration: vi.fn(),
  beginPasskeyLogin: vi.fn(),
  encodeAttestationResponse: vi.fn(),
  encodeAssertionResponse: vi.fn(),
}))

import { statusService } from '@/lib/services/status'
import { Route } from './login'

const mockStatus = statusService.get as ReturnType<typeof vi.fn>

/** Minimal status payload — only the fields the login page reads. */
function status(overrides: Partial<StatusInfo>): StatusInfo {
  return {
    google_oauth: false,
    google_client_id: '',
    github_oauth: false,
    github_client_id: '',
    passkey_login: false,
    ...overrides,
  } as StatusInfo
}

function renderLogin() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const LoginPage = (Route as unknown as { options: { component: React.ComponentType } }).options
    .component
  return render(
    <QueryClientProvider client={qc}>
      <LoginPage />
    </QueryClientProvider>,
  )
}

beforeEach(() => {
  vi.clearAllMocks()
})

describe('login page passkey gate', () => {
  it('renders "Sign in with Passkey" when status.passkey_login is true', async () => {
    mockStatus.mockResolvedValue(status({ passkey_login: true }))

    renderLogin()

    expect(await screen.findByText('Sign in with Passkey')).toBeInTheDocument()
  })

  it('hides the passkey button when status.passkey_login is false', async () => {
    // Google enabled so we can detect when the status payload has been applied.
    mockStatus.mockResolvedValue(
      status({ passkey_login: false, google_oauth: true, google_client_id: 'cid' }),
    )

    renderLogin()

    // Status has loaded once the OAuth button shows up …
    expect(await screen.findByText('Continue with Google')).toBeInTheDocument()
    // … and the passkey entry point is still absent.
    expect(screen.queryByText('Sign in with Passkey')).not.toBeInTheDocument()
  })
})
