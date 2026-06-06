package system_setting

import (
	"testing"

	"github.com/stretchr/testify/require"

	settingconfig "github.com/songquanpeng/one-api/setting/config"
)

// resetSettings clears state between tests since the singleton bleeds
// across t.Run boundaries. Don't t.Parallel any of these.
func resetSettings(t *testing.T) {
	t.Helper()
	defaultOIDCSettings = OIDCSettings{Providers: []OIDCProvider{}}
}

func TestGetOIDCProvider_HitAndMiss(t *testing.T) {
	resetSettings(t)
	defaultOIDCSettings.Providers = []OIDCProvider{
		{Name: "google", DisplayName: "Google", Enabled: true},
		{Name: "okta", DisplayName: "Okta", Enabled: false},
	}
	require.NotNil(t, GetOIDCProvider("google"))
	require.Equal(t, "Google", GetOIDCProvider("google").DisplayName)
	require.NotNil(t, GetOIDCProvider("okta"))
	require.Nil(t, GetOIDCProvider("missing"))
}

func TestEnabledOIDCProviders_FilteringAndSecretStripping(t *testing.T) {
	resetSettings(t)
	defaultOIDCSettings.Providers = []OIDCProvider{
		{Name: "g", DisplayName: "Google", Icon: "g.png", Enabled: true,
			ClientId: "id1", ClientSecret: "shh"},
		{Name: "o", DisplayName: "Okta", Enabled: false,
			ClientId: "id2", ClientSecret: "shh2"},
	}
	enabled := EnabledOIDCProviders()
	require.Len(t, enabled, 1)
	require.Equal(t, "g", enabled[0].Name)
	require.Equal(t, "Google", enabled[0].DisplayName)
	require.Equal(t, "g.png", enabled[0].Icon)
	// Secrets must NOT leak to /api/status.
	require.Empty(t, enabled[0].ClientId, "client_id must not be exposed")
	require.Empty(t, enabled[0].ClientSecret, "client_secret must not be exposed")
}

// Round-trip via the registry — proves slice-of-struct + json tags survive
// the LoadFromDB / SaveToDB cycle. Catches regressions if the reflection
// path in setting/config ever changes.
func TestOIDCSettings_RoundTripThroughRegistry(t *testing.T) {
	resetSettings(t)

	// Snapshot what an admin would PUT to /api/option/ to set up two IdPs.
	dbRows := map[string]string{
		"oidc.providers": `[{"name":"google","display_name":"Sign in with Google",` +
			`"well_known":"https://accounts.google.com/.well-known/openid-configuration",` +
			`"client_id":"gcid","client_secret":"gcs","enabled":true},` +
			`{"name":"okta","display_name":"Okta","well_known":"https://oktatest.com/.well-known/openid-configuration",` +
			`"client_id":"ocid","client_secret":"ocs","enabled":false}]`,
	}
	require.NoError(t, settingconfig.GlobalConfig.LoadFromDB(dbRows))

	require.Len(t, defaultOIDCSettings.Providers, 2)
	google := GetOIDCProvider("google")
	require.NotNil(t, google)
	require.Equal(t, "Sign in with Google", google.DisplayName)
	require.Equal(t, "gcid", google.ClientId)
	require.True(t, google.Enabled)

	okta := GetOIDCProvider("okta")
	require.NotNil(t, okta)
	require.False(t, okta.Enabled)
}

// Empty providers list must serialise/deserialise as [] (not null). Caught
// once that reflect.New + json.Unmarshal can produce nil slice if not
// initialised — the LoadFromDB path replaces wholesale so a zero-length
// JSON array stays as length zero.
func TestOIDCSettings_EmptyProvidersStaysEmpty(t *testing.T) {
	resetSettings(t)
	defaultOIDCSettings.Providers = []OIDCProvider{
		{Name: "stale", Enabled: true},
	}
	require.NoError(t, settingconfig.GlobalConfig.LoadFromDB(map[string]string{
		"oidc.providers": "[]",
	}))
	require.Len(t, defaultOIDCSettings.Providers, 0)
	require.Nil(t, GetOIDCProvider("stale"))
}
