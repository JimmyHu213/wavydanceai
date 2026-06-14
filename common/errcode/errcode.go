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
	AuthLoginDisabled        = "auth.login_disabled"
	AuthInvalidCredentials   = "auth.invalid_credentials"
	AuthRegisterDisabled     = "auth.register_disabled"
	AuthPasswordComplexity   = "auth.password_complexity"
	AuthVerificationRequired = "auth.verification_required"
	AuthVerificationFailed   = "auth.verification_failed"
	AuthResetLinkInvalid     = "auth.reset_link_invalid"

	// user
	UserUsernameTaken       = "user.username_taken"
	UserEmailTaken          = "user.email_taken"
	UserCreateFailed        = "user.create_failed"
	UserEmailNotRegistered  = "user.email_not_registered"
	UserPasswordResetFailed = "user.password_reset_failed"

	// email
	EmailDomainNotAllowed = "email.domain_not_allowed"
	EmailSendFailed       = "email.send_failed"
)

// All returns every defined code; used by tests to guard uniqueness/format.
func All() []string {
	return []string{
		ParamInvalid,
		ServerInternal,
		ServerSessionSaveFailed,
		AuthLoginDisabled,
		AuthInvalidCredentials,
		AuthRegisterDisabled,
		AuthPasswordComplexity,
		AuthVerificationRequired,
		AuthVerificationFailed,
		AuthResetLinkInvalid,
		UserUsernameTaken,
		UserEmailTaken,
		UserCreateFailed,
		UserEmailNotRegistered,
		UserPasswordResetFailed,
		EmailDomainNotAllowed,
		EmailSendFailed,
	}
}
