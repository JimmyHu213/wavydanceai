// Username constraints — mirror `validate:"max=12"` on `model.User.Username`
// so the form can fail fast instead of asking the backend to reject submissions
// after the user already filled in the rest of the registration fields.
//
// Length is counted in Unicode code points so it matches the rune-based length
// go-playground/validator and Go's `utf8.RuneCountInString` both use.
export const USERNAME_MIN = 3
export const USERNAME_MAX = 12

export type UsernameIssue = 'too_short' | 'too_long'

function codePointLength(s: string): number {
  return [...s].length
}

export function checkUsername(name: string): UsernameIssue | null {
  const n = codePointLength(name)
  if (n < USERNAME_MIN) return 'too_short'
  if (n > USERNAME_MAX) return 'too_long'
  return null
}
