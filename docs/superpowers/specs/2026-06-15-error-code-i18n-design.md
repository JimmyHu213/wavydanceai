# Error code-ification + frontend i18n mapping — design

**Date:** 2026-06-15
**Status:** Approved (design), pending implementation plan
**Branch:** `feat/error-code-i18n`

## Problem

The backend emits API errors as bare strings (mostly hardcoded Chinese) in the
`{success:false, message}` envelope. The frontend displays `message` verbatim.
Result: a user on the English UI still sees Chinese error text — error messages
are the one part of the product i18n does not cover. The only existing
workaround is a brittle exact-string match (`PasskeyCard`: `msg === 'passkey
disabled' ? t(...) : msg`) that breaks the moment the backend reworks a string.

## Goal

Make API errors render in the user's language. The mechanism: the backend
attaches a **stable error `code`** to each error; the frontend maps the code to
a localized string via its existing i18n catalog, falling back to the backend
`message` when the code is unknown.

`code` is not a separate deliverable — it is the minimum stable identifier that
makes translation robust (matching on Chinese strings is too fragile).

### Non-goals

- **Relay `/v1/**` errors** (OpenAI-compatible `{"error":{message,type,code}}`).
  That shape is an external contract consumed by SDKs; untouched.
- **HTTP status codes.** Many `/api` errors return `200` with `success:false`
  (an upstream one-api convention). Reshuffling status codes is risky and
  orthogonal; this work only *adds* a `code` field, it does not change which
  status each site returns.
- **Migrating all ~217 backend error sites.** Only the high-frequency,
  user-reachable flows are migrated now (see Scope). The architecture supports
  migrating the rest incrementally with no breaking change.
- **Replacing the backend `common/i18n` system.** It stays for the few places
  that use it; it is not expanded to carry all error copy.

## Architecture

Backend sends `code` + a default `message`; the frontend owns localization.

```
backend error site ──► SendError(c, status, errcode.X, "默认中文")
                         │
                         ▼
   response envelope:  {success:false, code:"param.invalid", message:"默认中文", data:null}
                         │
                         ▼
frontend api.ts ──► ApiError { code, message, status }
                         │
                         ▼
   errorText(e, t) ──► i18n has "errors.param.invalid"?  yes ─► localized string
                                                          no  ─► e.message (backend default)
                                                          (no code at all) ─► e.message / generic fallback
```

Unmigrated sites send no `code`; the frontend falls through to `message` exactly
as today. Migration is therefore strictly additive and can stop/resume at any
site without breaking anything.

## Components

### 1. Response envelope (backward compatible)

Add one optional field. Before:

```json
{ "success": false, "message": "参数错误", "data": null }
```

After (migrated site):

```json
{ "success": false, "code": "param.invalid", "message": "参数错误", "data": null }
```

`message` is always present (curl/API-direct consumers, unmigrated sites, and
the frontend fallback all rely on it). `code` is optional and omitted by
unmigrated sites.

### 2. Backend — `common/errcode/errcode.go`

A flat set of string constants, dotted-lowercase, grouped by domain. Format
mirrors the frontend i18n key structure so `code` → `errors.<code>` is mechanical.

```go
package errcode

const (
    ParamInvalid          = "param.invalid"
    AuthInvalidCredentials = "auth.invalid_credentials"
    AuthRateLimited        = "auth.rate_limited"
    UserEmailTaken         = "user.email_taken"
    UserNotFound           = "user.not_found"
    QuotaInsufficient      = "quota.insufficient"
    ChannelNotFound        = "channel.not_found"
    TokenNotFound          = "token.not_found"
    // …added as flows are migrated
)
```

No registration/init machinery (unlike `setting/`) — codes are compile-time
constants referenced by call sites and matched by frontend keys. The package is
the single source of truth for the code vocabulary.

### 3. Backend — `controller/response.go`

One helper, since no shared response helper exists today:

```go
// SendError writes the standard error envelope with a stable code.
// message is the human-readable default (kept as-is from the migrated site).
func SendError(c *gin.Context, httpStatus int, code, message string) {
    c.JSON(httpStatus, gin.H{"success": false, "code": code, "message": message})
}
```

Migrating a site is a one-line swap that preserves its status and message:

```go
// before
c.JSON(http.StatusOK, gin.H{"success": false, "message": "参数错误：" + err.Error()})
// after
SendError(c, http.StatusOK, errcode.ParamInvalid, "参数错误："+err.Error())
```

A matching `SendSuccess` is **not** in scope — success responses are untouched.

### 4. Frontend — `ApiError.code` + `errorText`

- `web/wavy/src/lib/api.ts`: `ApiError` gains `code?: string`; `unwrap` and the
  axios error interceptor read `res.data.code` alongside `message`.
- New util `web/wavy/src/lib/errorText.ts`:

```ts
export function errorText(e: unknown, t: TFunction, fallback?: string): string {
  if (e instanceof ApiError) {
    if (e.code && i18n.exists(`errors.${e.code}`)) return t(`errors.${e.code}`)
    if (e.message) return e.message
  }
  return fallback ?? t('errors.generic')
}
```

Migrated display sites replace `e instanceof ApiError ? e.message : t('...')`
with `errorText(e, t, t('...'))`. The ~116 existing sites keep working unchanged
and adopt `errorText` opportunistically.

### 5. Frontend i18n keys

Add an `errors` namespace to `web/wavy/src/locales/en.json` and `zh-CN.json`,
keyed to match the backend codes:

```json
"errors": {
  "generic": "Something went wrong. Please try again.",
  "param": { "invalid": "Invalid input." },
  "auth": { "invalid_credentials": "Incorrect username or password.", "rate_limited": "Too many attempts. Try again later." },
  "user": { "email_taken": "That email is already registered." },
  "quota": { "insufficient": "Insufficient balance. Please top up." }
}
```

(react-i18next reads dotted keys as nested objects, so `errors.auth.invalid_credentials` resolves.)

## Scope — flows migrated now

User-reachable, high-frequency paths and their controllers:

| Flow | Controller(s) |
| --- | --- |
| Login / register / password reset | `controller/user.go`, `controller/auth/*` |
| Top-up loop (redeem, stripe, epay, crypto) | `controller/topup.go` / top-up controllers |
| Token CRUD (create/update/delete) | `controller/token.go` |
| Channel CRUD (create/update/delete/test) | `controller/channel.go` |
| Billing / pricing save | `controller/option.go` (+ batch option endpoint) |
| Playground token + model lists | `controller/misc.go` / `controller/user.go` playground handlers |

Each migrated error site: pick or add a code in `errcode`, swap to `SendError`,
add the matching `errors.<code>` key in both locales, and point the relevant
frontend display site(s) at `errorText`.

Admin-only deep-corner errors (e.g. obscure channel-balance provider paths) are
left emitting legacy `message`; they display in the backend default language
until migrated later.

## Error handling / edge cases

- **Unknown code on frontend** (backend sends a code the locale doesn't have a
  key for): `i18n.exists` guard → falls back to `message`. No crash, no missing
  `errors.x.y` literal shown.
- **No code** (unmigrated site): falls back to `message`, i.e. today's behavior.
- **Neither code nor message**: generic fallback `errors.generic`.
- **API-direct / curl consumers**: still get a readable `message`; `code` is
  additive metadata they can ignore or adopt.

## Testing

- **Backend (Go):** `SendError` writes `{success:false, code, message}` at the
  given status; table test. For 2–3 migrated controllers, assert the response
  carries the expected `code` (e.g. bad login → `auth.invalid_credentials`).
- **Frontend (Vitest):** `errorText` three branches — known code → localized,
  unknown/absent code → `message`, neither → generic fallback. Plus 1–2 migrated
  flows asserting the localized string renders (mock service rejects with a
  coded `ApiError`).
- Full suites green: `go test ./...`, `cd web/wavy && bun run test && bun run build`.

## Backward compatibility

Strictly additive. Envelope gains an optional field; unmigrated backend sites
and the ~110 not-yet-touched frontend display sites behave exactly as before.
No endpoint contract is broken; relay `/v1` untouched.

## Risks

- **Code vocabulary churn:** picking code names ad hoc could fragment. Mitigation:
  all codes live in `common/errcode`; reviewers reject inline string codes.
- **Locale drift:** a code added backend-side without the matching frontend key
  silently falls back to `message`. Acceptable (no breakage), but the migrated
  flows' tests pin the keys that matter.
- **Scope creep into status-code fixes:** explicitly out of scope; reviewers
  keep each migration a pure `code`-addition.
