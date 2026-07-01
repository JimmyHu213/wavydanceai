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

	// token
	TokenNotFound   = "token.not_found"
	TokenExpired    = "token.expired"
	TokenExhausted  = "token.exhausted"
	TokenSaveFailed = "token.save_failed"

	// channel
	ChannelNotFound   = "channel.not_found"
	ChannelSaveFailed = "channel.save_failed"
	ChannelTestFailed = "channel.test_failed"

	// topup
	TopupPaymentsDisabled   = "topup.payments_disabled"
	TopupGatewayUnavailable = "topup.gateway_unavailable"
	TopupAmountBelowMinimum = "topup.amount_below_minimum"
	TopupRedeemFailed       = "topup.redeem_failed"

	// option
	OptionInvalidValue = "option.invalid_value"
	OptionSaveFailed   = "option.save_failed"

	// playground
	PlaygroundTokenFailed  = "playground.token_failed"
	PlaygroundModelsFailed = "playground.models_failed"

	// auth (cont.)
	AuthUnauthenticated = "auth.unauthenticated"
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
		TokenNotFound,
		TokenExpired,
		TokenExhausted,
		TokenSaveFailed,
		ChannelNotFound,
		ChannelSaveFailed,
		ChannelTestFailed,
		TopupPaymentsDisabled,
		TopupGatewayUnavailable,
		TopupAmountBelowMinimum,
		TopupRedeemFailed,
		OptionInvalidValue,
		OptionSaveFailed,
		PlaygroundTokenFailed,
		PlaygroundModelsFailed,
		AuthUnauthenticated,
	}
}
