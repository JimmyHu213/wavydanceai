// Package system_setting holds settings for cross-cutting platform
// concerns: auth IdPs, system theme, legal pages, etc. Each file registers
// exactly one settings struct via setting/config; admins edit via
// /api/option/ using keys formatted as "<module>.<json_tag>".
//
// Naming aligned with QuantumNous/new-api: struct is plural (OIDCSettings,
// not OIDCSetting); register key drops the "_setting" suffix ("oidc", not
// "oidc_setting"); subpackage is system_setting/ not auth_setting/.
package system_setting

import (
	"github.com/songquanpeng/one-api/setting/config"
)

// OIDCProvider is one configured IdP that speaks OpenID Connect. Endpoints
// are derived at runtime from WellKnown via service/oauth/oidc/discovery —
// only WellKnown + credentials + display metadata are stored.
//
// Name must be a URL-safe slug (used as the :provider path parameter on
// /api/oauth/oidc/:provider). It's also prefixed onto User.OidcId as
// "<name>:<sub>" so two providers can hand out the same `sub` without
// colliding on a single user record.
type OIDCProvider struct {
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	Icon         string `json:"icon,omitempty"`
	WellKnown    string `json:"well_known"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Enabled      bool   `json:"enabled"`
}

// OIDCSettings is the registry. Stored as one row in the option table
// (key: "oidc.providers") JSON-encoded — the reflection-based persistence
// in setting/config handles slice-of-struct round trip.
type OIDCSettings struct {
	Providers []OIDCProvider `json:"providers"`
}

var defaultOIDCSettings = OIDCSettings{
	Providers: []OIDCProvider{},
}

func init() {
	config.GlobalConfig.Register("oidc", &defaultOIDCSettings)
}

// GetOIDCSettings returns the live pointer — never copy the struct or
// you'll miss subsequent admin updates.
func GetOIDCSettings() *OIDCSettings {
	return &defaultOIDCSettings
}

// GetOIDCProvider returns the provider with the matching Name, or nil. The
// returned pointer is read-only — mutating it bypasses the registry's
// persistence path.
func GetOIDCProvider(name string) *OIDCProvider {
	for i := range defaultOIDCSettings.Providers {
		if defaultOIDCSettings.Providers[i].Name == name {
			return &defaultOIDCSettings.Providers[i]
		}
	}
	return nil
}

// EnabledOIDCProviders returns the subset that admins have switched on.
// Used by /api/status to drive the login UI's button list. Secrets are
// stripped — only Name / DisplayName / Icon are safe to expose.
func EnabledOIDCProviders() []OIDCProvider {
	out := []OIDCProvider{}
	for _, p := range defaultOIDCSettings.Providers {
		if !p.Enabled {
			continue
		}
		out = append(out, OIDCProvider{
			Name:        p.Name,
			DisplayName: p.DisplayName,
			Icon:        p.Icon,
			Enabled:     true,
		})
	}
	return out
}
