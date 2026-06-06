// Package passkey holds runtime configuration for WebAuthn / Passkey login.
// Admins edit these via /api/option/ with keys "passkey_setting.<json_tag>".
package passkey

import (
	"github.com/songquanpeng/one-api/setting/config"
)

// PasskeySetting controls the Passkey login feature. When Enabled is false
// all /passkey/* and /login/passkey/* endpoints reject with HTTP 403, so the
// table can ship dark and be toggled on per environment.
type PasskeySetting struct {
	Enabled   bool   `json:"enabled"`
	RPID      string `json:"rp_id"`
	RPName    string `json:"rp_name"`
	RPOrigins string `json:"rp_origins"` // JSON array of origins, e.g. ["https://wavydance.ai"]
}

var passkeySetting = PasskeySetting{
	Enabled:   false,
	RPID:      "",
	RPName:    "",
	RPOrigins: "",
}

func init() {
	config.GlobalConfig.Register("passkey_setting", &passkeySetting)
}

// GetPasskeySetting returns the live pointer; never copy.
func GetPasskeySetting() *PasskeySetting {
	return &passkeySetting
}
