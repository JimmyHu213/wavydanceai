import { describe, it, expect, afterEach } from 'vitest'
import { AxiosError, type AxiosAdapter, type AxiosResponse } from 'axios'
import { api } from './api'

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
  it('replaces the axios default message with the backend business message', async () => {
    api.defaults.adapter = respondWith(403, { success: false, message: 'passkey disabled' })

    await expect(api.get('/user/passkey/credentials')).rejects.toMatchObject({
      message: 'passkey disabled',
    })
  })

  it('keeps the axios default message when the body has no message', async () => {
    api.defaults.adapter = respondWith(500, { success: false })

    await expect(api.get('/anything')).rejects.toMatchObject({
      message: 'Request failed with status code 500',
    })
  })
})
