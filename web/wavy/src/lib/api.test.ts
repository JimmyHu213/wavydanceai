import { describe, it, expect, afterEach } from 'vitest'
import { AxiosError, type AxiosAdapter, type AxiosResponse } from 'axios'
import { api, ApiError, unwrap } from './api'

/** Fake adapter, no network involved. Non-2xx rejects with the same
 *  AxiosError shape (default message + response) that axios's settle()
 *  produces, which exercises the real response error interceptor. */
function respondWith(status: number, data: unknown): AxiosAdapter {
  return async (config) => {
    const response = { status, statusText: '', headers: {}, config, data } as AxiosResponse
    if (status >= 200 && status < 300) return response
    throw new AxiosError(
      `Request failed with status code ${status}`,
      AxiosError.ERR_BAD_REQUEST,
      config,
      null,
      response,
    )
  }
}

const originalAdapter = api.defaults.adapter

afterEach(() => {
  api.defaults.adapter = originalAdapter
})

describe('api response error interceptor', () => {
  it('rejects an ApiError carrying the backend business message and status', async () => {
    api.defaults.adapter = respondWith(403, { success: false, message: 'passkey disabled' })

    const err = await api.get('/user/passkey/credentials').catch((e: unknown) => e)

    // instanceof matters: ~20 call sites gate on `e instanceof ApiError ? e.message : fallback`.
    expect(err).toBeInstanceOf(ApiError)
    expect(err).toMatchObject({ message: 'passkey disabled', status: 403 })
  })

  it('keeps the axios default message (still as ApiError) when the body has no message', async () => {
    api.defaults.adapter = respondWith(500, { success: false })

    const err = await api.get('/anything').catch((e: unknown) => e)

    expect(err).toBeInstanceOf(ApiError)
    expect(err).toMatchObject({ message: 'Request failed with status code 500', status: 500 })
  })
})

describe('unwrap', () => {
  it('throws ApiError carrying the backend code', () => {
    const res = {
      data: { success: false, code: 'auth.invalid_credentials', message: '用户名或密码错误' },
      status: 200,
    } as Parameters<typeof unwrap>[0]
    try {
      unwrap(res)
      expect.unreachable('should have thrown')
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError)
      expect((e as ApiError).code).toBe('auth.invalid_credentials')
      expect((e as ApiError).message).toBe('用户名或密码错误')
    }
  })

  it('leaves code undefined when the envelope has none', () => {
    const res = { data: { success: false, message: 'oops' }, status: 200 } as Parameters<typeof unwrap>[0]
    try {
      unwrap(res)
    } catch (e) {
      expect((e as ApiError).code).toBeUndefined()
    }
  })
})
