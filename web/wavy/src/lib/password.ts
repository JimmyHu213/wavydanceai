// Password complexity rule — mirrors `common.IsPasswordComplexEnough` on the
// backend so the UI can fail fast with the same constraint.
export const PASSWORD_MIN = 8
export const PASSWORD_MAX = 24

const HAS_LETTER = /\p{L}/u
const HAS_DIGIT = /\p{N}/u

export type PasswordIssue = 'too_short' | 'too_long' | 'needs_letter_and_digit'

export function checkPassword(pw: string): PasswordIssue | null {
  if (pw.length < PASSWORD_MIN) return 'too_short'
  if (pw.length > PASSWORD_MAX) return 'too_long'
  if (!HAS_LETTER.test(pw) || !HAS_DIGIT.test(pw)) return 'needs_letter_and_digit'
  return null
}
