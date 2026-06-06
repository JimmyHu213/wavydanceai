package passkey

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/songquanpeng/one-api/setting/config"
)

func TestPasskeySettingRoundTrip(t *testing.T) {
	// Registry round-trip: export → load → fields restored.
	s := GetPasskeySetting()
	prev := *s
	t.Cleanup(func() { *s = prev })

	s.Enabled = true
	s.RPID = "wavydance.ai"
	s.RPName = "Wavy Dance AI"
	s.RPOrigins = `["https://wavydance.ai"]`

	exported := config.GlobalConfig.ExportAllConfigs()
	require.Equal(t, "true", exported["passkey_setting.enabled"])
	require.Equal(t, "wavydance.ai", exported["passkey_setting.rp_id"])
	require.Equal(t, `["https://wavydance.ai"]`, exported["passkey_setting.rp_origins"])

	*s = PasskeySetting{}
	require.NoError(t, config.GlobalConfig.LoadFromDB(map[string]string{
		"passkey_setting.enabled":    "true",
		"passkey_setting.rp_id":      "wavydance.ai",
		"passkey_setting.rp_name":    "Wavy Dance AI",
		"passkey_setting.rp_origins": `["https://wavydance.ai"]`,
	}))
	require.True(t, s.Enabled)
	require.Equal(t, "wavydance.ai", s.RPID)
}
