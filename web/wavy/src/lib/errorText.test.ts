import { describe, it, expect } from 'vitest'
import { ApiError } from './api'
import { errorText } from './errorText'

function makeI18n(keys: Record<string, string>) {
  return { exists: (k: string) => k in keys, t: (k: string) => keys[k] ?? k }
}

describe('errorText', () => {
  it('translates a known code', () => {
    const i = makeI18n({ 'errors.auth.invalid_credentials': 'Incorrect username or password.' })
    const e = new ApiError('用户名或密码错误', 200, 'auth.invalid_credentials')
    expect(errorText(e, i.t as never, 'fallback', i as never)).toBe('Incorrect username or password.')
  })

  it('falls back to backend message for an unknown code', () => {
    const i = makeI18n({})
    const e = new ApiError('后端原文', 200, 'auth.invalid_credentials')
    expect(errorText(e, i.t as never, 'fallback', i as never)).toBe('后端原文')
  })

  it('uses the provided fallback when not an ApiError', () => {
    const i = makeI18n({})
    expect(errorText(new Error('x'), i.t as never, 'fallback', i as never)).toBe('fallback')
  })
})
