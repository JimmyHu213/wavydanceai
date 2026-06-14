# Error code-ification + frontend i18n mapping — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `/api` error messages render in the user's language by tagging each migrated backend error with a stable `code` and translating it on the frontend, falling back to the backend `message` when the code is unknown.

**Architecture:** Backend gains an `errcode` constant package + a `SendError` helper that writes `{success:false, code, message}`. Migrated controller sites swap their hand-rolled `c.JSON(gin.H{...})` for `SendError`, keeping the existing status and message. Frontend `ApiError` carries `code`; a new `errorText(e, t, fallback)` util translates `errors.<code>` when present, else shows `message`. Changes are strictly additive — unmigrated sites and untouched frontend sites behave exactly as today.

**Tech Stack:** Go (gin), Vitest + React Testing Library, react-i18next.

**Spec:** `docs/superpowers/specs/2026-06-15-error-code-i18n-design.md`

---

## File Structure

**Backend (create):**
- `common/errcode/errcode.go` — string-constant vocabulary of error codes (single source of truth)
- `common/errcode/errcode_test.go` — uniqueness + format guard
- `controller/response.go` — `SendError` helper
- `controller/response_test.go` — helper behavior

**Backend (modify):**
- `controller/user.go` — Login / Register / reset sites → `SendError`
- `controller/token.go` — token CRUD sites
- `controller/channel.go` — channel CRUD sites
- `controller/topup.go` (+ sibling top-up controllers) — top-up loop sites
- `controller/option.go` — billing/option save sites
- playground handler (`controller/misc.go` or `controller/user.go`) — playground token/model sites

**Frontend (create):**
- `web/wavy/src/lib/errorText.ts` — code→i18n resolver with fallback
- `web/wavy/src/lib/errorText.test.ts` — three-branch behavior

**Frontend (modify):**
- `web/wavy/src/lib/types.ts` — `ApiResponse.code?`
- `web/wavy/src/lib/api.ts` — `ApiError.code`; `unwrap` + interceptor read it
- `web/wavy/src/locales/en.json` + `zh-CN.json` — `errors` namespace
- `web/wavy/src/routes/login.tsx` (and the other migrated flows' display sites) — use `errorText`

---

## Task 1: Backend errcode package

**Files:**
- Create: `common/errcode/errcode.go`
- Test: `common/errcode/errcode_test.go`

- [ ] **Step 1: Write the failing test**

`common/errcode/errcode_test.go`:

```go
package errcode

import (
	"regexp"
	"testing"
)

// Codes are a wire contract with the frontend i18n keys. Guard against
// accidental duplicate values (copy-paste) and malformed names.
func TestCodesAreUniqueAndWellFormed(t *testing.T) {
	codes := All()
	if len(codes) == 0 {
		t.Fatal("All() returned no codes")
	}
	format := regexp.MustCompile(`^[a-z]+(\.[a-z_]+)+$`)
	seen := map[string]bool{}
	for _, c := range codes {
		if !format.MatchString(c) {
			t.Errorf("code %q is not dotted-lowercase (e.g. domain.reason)", c)
		}
		if seen[c] {
			t.Errorf("duplicate code value %q", c)
		}
		seen[c] = true
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/jimmy/Documents/Projects/wavydanceai-feat-error-i18n && go test ./common/errcode/`
Expected: FAIL — package/`All` undefined.

- [ ] **Step 3: Write minimal implementation**

`common/errcode/errcode.go`:

```go
// Package errcode is the single source of truth for /api error codes.
// Each code is a stable, dotted-lowercase identifier (domain.reason) that the
// frontend maps to a localized string via the `errors.<code>` i18n key. The
// human-readable message stays at the call site as a fallback.
package errcode

const (
	// generic / server
	ParamInvalid            = "param.invalid"
	ServerInternal          = "server.internal"
	ServerSessionSaveFailed = "server.session_save_failed"

	// auth
	AuthLoginDisabled      = "auth.login_disabled"
	AuthInvalidCredentials = "auth.invalid_credentials"
)

// All returns every defined code; used by tests to guard uniqueness/format.
func All() []string {
	return []string{
		ParamInvalid,
		ServerInternal,
		ServerSessionSaveFailed,
		AuthLoginDisabled,
		AuthInvalidCredentials,
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./common/errcode/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add common/errcode/
git commit -m "feat(errcode): error code vocabulary package"
```

---

## Task 2: Backend SendError helper

**Files:**
- Create: `controller/response.go`
- Test: `controller/response_test.go`

- [ ] **Step 1: Write the failing test**

`controller/response_test.go`:

```go
package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/errcode"
)

func TestSendError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	SendError(c, http.StatusOK, errcode.ParamInvalid, "参数错误")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body struct {
		Success bool   `json:"success"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Success {
		t.Error("success = true, want false")
	}
	if body.Code != errcode.ParamInvalid {
		t.Errorf("code = %q, want %q", body.Code, errcode.ParamInvalid)
	}
	if body.Message != "参数错误" {
		t.Errorf("message = %q, want 参数错误", body.Message)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./controller/ -run TestSendError`
Expected: FAIL — `SendError` undefined.

- [ ] **Step 3: Write minimal implementation**

`controller/response.go`:

```go
package controller

import "github.com/gin-gonic/gin"

// SendError writes the standard error envelope with a stable code.
// code comes from common/errcode; message is the human-readable default the
// frontend falls back to when it has no translation for the code. httpStatus
// keeps whatever the original site used — this helper does not change status
// semantics.
func SendError(c *gin.Context, httpStatus int, code, message string) {
	c.JSON(httpStatus, gin.H{
		"success": false,
		"code":    code,
		"message": message,
	})
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./controller/ -run TestSendError`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add controller/response.go controller/response_test.go
git commit -m "feat(controller): SendError helper for coded error envelope"
```

---

## Task 3: Migrate the Login flow (backend)

**Files:**
- Modify: `controller/user.go` (Login func, lines ~28-93)
- Modify: `common/errcode/errcode.go` (codes already defined in Task 1 cover this flow)
- Test: `controller/user_login_test.go` (create)

This proves the backend half of the vertical slice. Login's own `c.JSON` error
sites become `SendError`. The `ValidateAndFill` failure (bad creds / banned —
already merged into one string in `model/user.go:214,219`) maps to
`auth.invalid_credentials`.

- [ ] **Step 1: Write the failing test**

`controller/user_login_test.go`:

```go
package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/errcode"
)

// Login with password login disabled returns the coded envelope.
func TestLogin_PasswordLoginDisabled_HasCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prev := config.PasswordLoginEnabled
	config.PasswordLoginEnabled = false
	defer func() { config.PasswordLoginEnabled = prev }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/user/login",
		strings.NewReader(`{"username":"x","password":"y"}`))

	Login(c)

	var body struct {
		Success bool   `json:"success"`
		Code    string `json:"code"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body.Success || body.Code != errcode.AuthLoginDisabled {
		t.Fatalf("got success=%v code=%q, want false / %q", body.Success, body.Code, errcode.AuthLoginDisabled)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./controller/ -run TestLogin_PasswordLoginDisabled_HasCode`
Expected: FAIL — response has no `code` field yet.

- [ ] **Step 3: Migrate the Login error sites**

In `controller/user.go`, replace the error `c.JSON` sites inside `Login` (and `SetupLogin`'s session-save site). Exact swaps:

```go
// "管理员关闭了密码登录"  (was lines ~29-32)
SendError(c, http.StatusOK, errcode.AuthLoginDisabled, "管理员关闭了密码登录")

// decode failure (was ~38-41)
SendError(c, http.StatusOK, errcode.ParamInvalid, i18n.Translate(c, "invalid_parameter"))

// empty username/password (was ~47-50)
SendError(c, http.StatusOK, errcode.ParamInvalid, i18n.Translate(c, "invalid_parameter"))

// ValidateAndFill failure (was ~59-62) — bad creds / banned
SendError(c, http.StatusOK, errcode.AuthInvalidCredentials, err.Error())

// session save failure inside the 2FA-pending branch (was ~73)
SendError(c, http.StatusOK, errcode.ServerSessionSaveFailed, err.Error())
```

Also in `SetupLogin` (the "无法保存会话信息，请重试" site, ~105):

```go
SendError(c, http.StatusOK, errcode.ServerSessionSaveFailed, "无法保存会话信息，请重试")
```

Leave the `success: true` responses untouched. Remove no imports unless they become unused.

- [ ] **Step 4: Run test + full controller build**

Run: `go test ./controller/ -run TestLogin && go build ./...`
Expected: PASS, build clean.

- [ ] **Step 5: Commit**

```bash
git add controller/user.go controller/user_login_test.go
git commit -m "feat(auth): code-ify Login error responses"
```

---

## Task 4: Frontend — ApiResponse.code + ApiError.code

**Files:**
- Modify: `web/wavy/src/lib/types.ts:3-7`
- Modify: `web/wavy/src/lib/api.ts` (ApiError class, interceptor, unwrap)
- Test: `web/wavy/src/lib/api.test.ts` (create)

- [ ] **Step 1: Write the failing test**

`web/wavy/src/lib/api.test.ts`:

```ts
import { describe, it, expect } from 'vitest'
import { ApiError, unwrap } from './api'

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
    const res = { data: { success: false, message: 'oops' }, status: 200 } as Parameters<
      typeof unwrap
    >[0]
    try {
      unwrap(res)
    } catch (e) {
      expect((e as ApiError).code).toBeUndefined()
    }
  })
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd web/wavy && bun run test -- api.test`
Expected: FAIL — `ApiError` has no `code`.

- [ ] **Step 3: Implement**

In `web/wavy/src/lib/types.ts`, extend the envelope:

```ts
export interface ApiResponse<T = unknown> {
  success: boolean
  message: string
  code?: string
  data?: T
}
```

In `web/wavy/src/lib/api.ts`, add `code` to `ApiError` and read it in both throw paths:

```ts
export class ApiError extends Error {
  constructor(
    public message: string,
    public status?: number,
    public code?: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}
```

Interceptor reject (after computing `message`):

```ts
const code = typeof err.response?.data?.code === 'string' ? err.response.data.code : undefined
return Promise.reject(new ApiError(message, err.response?.status, code))
```

`unwrap`:

```ts
export function unwrap<T>(res: AxiosResponse<ApiResponse<T>>): T {
  if (!res.data.success)
    throw new ApiError(res.data.message || 'request failed', res.status, res.data.code)
  return res.data.data as T
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bun run test -- api.test`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/wavy/src/lib/types.ts web/wavy/src/lib/api.ts web/wavy/src/lib/api.test.ts
git commit -m "feat(web): ApiError carries backend error code"
```

---

## Task 5: Frontend — errorText resolver

**Files:**
- Create: `web/wavy/src/lib/errorText.ts`
- Test: `web/wavy/src/lib/errorText.test.ts`

- [ ] **Step 1: Write the failing test**

`web/wavy/src/lib/errorText.test.ts`:

```ts
import { describe, it, expect } from 'vitest'
import { ApiError } from './api'
import { errorText } from './errorText'

// Minimal stand-ins for i18next's instance + t.
function makeI18n(keys: Record<string, string>) {
  return {
    exists: (k: string) => k in keys,
    t: (k: string) => keys[k] ?? k,
  }
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `bun run test -- errorText`
Expected: FAIL — module missing.

- [ ] **Step 3: Implement**

`web/wavy/src/lib/errorText.ts`:

```ts
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `bun run test -- errorText`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add web/wavy/src/lib/errorText.ts web/wavy/src/lib/errorText.test.ts
git commit -m "feat(web): errorText resolver (code → i18n, message fallback)"
```

---

## Task 6: Frontend — errors.* keys + wire Login display (vertical slice complete)

**Files:**
- Modify: `web/wavy/src/locales/en.json`, `web/wavy/src/locales/zh-CN.json`
- Modify: `web/wavy/src/routes/login.tsx:90,131` (and the other login-factor catch sites)
- Test: `web/wavy/src/routes/login.test.tsx` (extend if present, else create a focused test)

- [ ] **Step 1: Add the errors namespace**

In `web/wavy/src/locales/en.json` (top-level object), add:

```json
"errors": {
  "generic": "Something went wrong. Please try again.",
  "param": { "invalid": "Invalid input." },
  "server": { "internal": "Server error. Please try again.", "session_save_failed": "Could not save your session. Please try again." },
  "auth": { "login_disabled": "Password login is disabled.", "invalid_credentials": "Incorrect username or password, or the account is disabled." }
}
```

In `web/wavy/src/locales/zh-CN.json`, the same keys with Chinese values:

```json
"errors": {
  "generic": "出错了，请重试。",
  "param": { "invalid": "参数有误。" },
  "server": { "internal": "服务器错误，请重试。", "session_save_failed": "无法保存会话信息，请重试。" },
  "auth": { "login_disabled": "管理员关闭了密码登录。", "invalid_credentials": "用户名或密码错误，或账号已被封禁。" }
}
```

- [ ] **Step 2: Write the failing test**

`web/wavy/src/routes/login.test.tsx` — assert a coded login failure renders the localized string, not the backend message. (Follow the existing route-test pattern in the repo, e.g. `console.pricing.test.tsx` / `console.tokens.test.tsx`: mock the auth service to reject with `new ApiError('后端原文', 200, 'auth.invalid_credentials')`, render the login route with providers, submit, and assert the English copy appears.)

```tsx
// key assertion after triggering a failed submit:
expect(await screen.findByText('Incorrect username or password, or the account is disabled.')).toBeInTheDocument()
```

- [ ] **Step 3: Run test to verify it fails**

Run: `bun run test -- login`
Expected: FAIL — raw backend message shown instead of localized text.

- [ ] **Step 4: Wire login.tsx to errorText**

Replace each `setErr(e instanceof ApiError ? e.message : t('login.failed'))` in `login.tsx` (lines ~90, ~131, and the other factor catch sites) with:

```tsx
setErr(errorText(e, t, t('login.failed')))
```

Add the import: `import { errorText } from '@/lib/errorText'`. Keep the existing `ApiError` import only if still referenced; otherwise remove it.

- [ ] **Step 5: Run test + build**

Run: `bun run test -- login && bun run build`
Expected: PASS, build clean.

- [ ] **Step 6: Commit**

```bash
git add web/wavy/src/locales/ web/wavy/src/routes/login.tsx web/wavy/src/routes/login.test.tsx
git commit -m "feat(web): localize login errors via errorText + errors.* keys"
```

---

## Tasks 7-11: Migrate remaining high-frequency flows

Each flow repeats the proven pattern from Tasks 3 + 6. Per flow: (a) add any new
codes to `common/errcode/errcode.go` **and its `All()` slice**, (b) swap that
controller's user-reachable error `c.JSON` sites to `SendError`, (c) add a test
asserting one representative coded error, (d) add matching `errors.<code>` keys
to both locales, (e) point the flow's frontend display site(s) at `errorText`,
(f) commit. Run `go build ./...`, `go test ./controller/...`, and
`cd web/wavy && bun run test && bun run build` before each commit.

Do these as separate commits so each is independently reviewable.

### Task 7: Register / password-reset (`controller/user.go` Register + reset, `controller/auth/*`)
New codes: `auth.register_disabled = "auth.register_disabled"`, `user.email_taken = "user.email_taken"`, `auth.verification_failed = "auth.verification_failed"`.
Frontend display sites: `web/wavy/src/routes/register.tsx`, password-reset route(s).

### Task 8: Token CRUD (`controller/token.go`)
New codes: `token.not_found = "token.not_found"`, `token.name_required = "token.name_required"` (as the actual sites dictate — read each `success:false` site and name by reason).
Frontend display sites: `web/wavy/src/routes/console.tokens.tsx` (already uses an onError from the silent-error PR — switch its message source to `errorText`).

### Task 9: Channel CRUD (`controller/channel.go`)
New codes: `channel.not_found = "channel.not_found"`, `channel.test_failed = "channel.test_failed"`.
Frontend display sites: `web/wavy/src/routes/console.channels.tsx`, `web/wavy/src/components/console/ChannelDialog.tsx`.

### Task 10: Top-up loop (`controller/topup.go` + sibling top-up controllers)
New codes: `topup.redeem_failed = "topup.redeem_failed"`, `topup.gateway_unavailable = "topup.gateway_unavailable"`, `quota.insufficient = "quota.insufficient"`.
Frontend display sites: `web/wavy/src/routes/console.topup.tsx`, `web/wavy/src/routes/topup-result.tsx`.

### Task 11: Billing/option save (`controller/option.go` + batch option endpoint) and playground (`controller/misc.go`/`controller/user.go` playground handlers)
New codes: `option.invalid_value = "option.invalid_value"`, `playground.token_failed = "playground.token_failed"`.
Frontend display sites: `web/wavy/src/components/console/pricing/PricingEditor.tsx`, `web/wavy/src/routes/console.playground.chat.tsx`, `web/wavy/src/components/playground/media/MediaPlayground.tsx`.

---

## Task 12: Final verification + PR

- [ ] **Step 1: Full backend suite**

Run: `go build ./... && go vet ./... && go test ./...`
Expected: all clean, 0 FAIL.

- [ ] **Step 2: Full frontend suite + build**

Run: `cd web/wavy && bun run test && bun run build`
Expected: all green.

- [ ] **Step 3: Grep guard — no inline string codes**

Run: `grep -rn '"code":' controller/ | grep -v errcode`
Expected: no matches (every code comes from `errcode`, not a literal).

- [ ] **Step 4: Open the PR**

```bash
git push -u origin feat/error-code-i18n
gh pr create --title "feat(api): code-ify high-frequency errors + frontend i18n mapping" --body "..."
```

PR body lists: the errcode package + SendError, the ApiError.code + errorText
resolver, the migrated flows (login/register/reset, token, channel, top-up,
billing/option, playground), the additive-envelope/backward-compat note, and
test coverage. Include the spec path. Do NOT merge; await CI + review.

---

## Self-Review

- **Spec coverage:** envelope `code` (Task 4) ✓; errcode package (Task 1) ✓; SendError (Task 2) ✓; ApiError.code (Task 4) ✓; errorText three-branch fallback (Task 5) ✓; errors.* keys (Task 6) ✓; migrated flows = login(3,6)/register-reset(7)/token(8)/channel(9)/topup(10)/billing+playground(11) ✓; relay & status codes untouched (no task touches relay/ or changes status) ✓; testing strategy (per-task tests + Task 12) ✓.
- **Placeholder scan:** Tasks 7-11 intentionally compress the *repeating* mechanical pattern but each names its exact files, the concrete codes to add, and the exact frontend display sites — they are executed by reading each `success:false` site in the named controller and applying the Task-3/Task-6 pattern shown in full above. The PR body `"..."` is filled at Task 12 step 4 from the listed contents.
- **Type consistency:** `SendError(c, httpStatus, code, message)` signature consistent across Tasks 2/3/7-11; `ApiError(message, status, code)` consistent across Tasks 4/5/6; `errorText(e, t, fallback, i18n?)` consistent across Tasks 5/6; `errcode.All()` updated in every code-adding task per the Task-7-11 preamble.
