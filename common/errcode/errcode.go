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
