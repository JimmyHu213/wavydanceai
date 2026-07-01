import type { TFunction } from 'i18next'
import i18nDefault from './i18n'
import { ApiError } from './api'

type I18nLike = { exists: (k: string) => boolean }

/**
 * Resolve a user-facing error string from a thrown value.
 * - ApiError with a code the catalog knows → localized `errors.<code>`
 * - ApiError otherwise → backend message (today's behavior)
 * - anything else → the caller's fallback, or generic
 * The `i18n` param is injectable for tests; defaults to the app instance.
 */
export function errorText(
  e: unknown,
  t: TFunction,
  fallback?: string,
  i18n: I18nLike = i18nDefault as unknown as I18nLike,
): string {
  if (e instanceof ApiError) {
    if (e.code && i18n.exists(`errors.${e.code}`)) return t(`errors.${e.code}`)
    if (e.message) return e.message
  }
  return fallback ?? t('errors.generic')
}
