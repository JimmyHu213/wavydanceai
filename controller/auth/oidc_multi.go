// Multi-provider OIDC handler. Mirrors controller/auth/oidc.go's flow but
// dispatches on a :provider path parameter and reads endpoints from the
// well-known discovery doc (cached in service/oauth/oidc) instead of
// per-provider flat config.
//
// User binding: User.OidcId stores "<provider_name>:<sub>" so two providers
// handing out the same `sub` don't collide. The single-provider legacy
// handler in oidc.go keeps writing the bare `sub` for backward compat —
// migrating those users to the prefixed form is a follow-up.
//
// Routes:
//   GET /api/oauth/oidc/:provider — login / register / bind
package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/controller"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/service/oauth/oidc"
	"github.com/songquanpeng/one-api/setting/system_setting"
)

// httpClient is reused across calls — cheaper than allocating per-request
// and gives us a single place to tune timeouts if Google/Okta/etc. ever go
// slow.
var oidcMultiClient = &http.Client{Timeout: 15 * time.Second}

// OIDCMultiAuth handles GET /api/oauth/oidc/:provider. State check + bind
// detection mirror the legacy handler. The interesting bits are the
// discovery lookup and the OidcId namespacing.
func OIDCMultiAuth(c *gin.Context) {
	ctx := c.Request.Context()
	session := sessions.Default(c)

	state := c.Query("state")
	if state == "" || session.Get("oauth_state") == nil || state != session.Get("oauth_state").(string) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "state is empty or not same",
		})
		return
	}

	providerName := c.Param("provider")
	provider := system_setting.GetOIDCProvider(providerName)
	if provider == nil || !provider.Enabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("OIDC provider %q 未配置或未启用", providerName),
		})
		return
	}

	discovery, err := oidc.Fetch(provider.WellKnown)
	if err != nil {
		logger.SysError("oidc multi discovery failed: " + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "OIDC 配置发现失败，请检查 well-known URL",
		})
		return
	}

	// Already-signed-in users sent through this endpoint are binding the
	// provider to their existing account.
	if session.Get("username") != nil {
		oidcMultiBind(c, provider, discovery)
		return
	}

	userInfo, err := exchangeOIDCMultiUserInfo(provider, discovery, c.Query("code"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	namespacedId := provider.Name + ":" + userInfo.OpenID
	user := model.User{OidcId: namespacedId}
	if model.IsOidcIdAlreadyTaken(namespacedId) {
		if err := user.FillUserByOidcId(); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			return
		}
	} else {
		if !config.RegisterEnabled {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "管理员关闭了新用户注册"})
			return
		}
		user.Email = userInfo.Email
		if userInfo.PreferredUsername != "" {
			user.Username = userInfo.PreferredUsername
		} else {
			user.Username = provider.Name + "_" + strconv.Itoa(model.GetMaxUserId()+1)
		}
		if userInfo.Name != "" {
			user.DisplayName = userInfo.Name
		} else {
			user.DisplayName = provider.DisplayName
		}
		if err := user.Insert(ctx, 0); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
			return
		}
	}

	if user.Status != model.UserStatusEnabled {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "用户已被封禁"})
		return
	}
	controller.SetupLogin(&user, c)
}

// oidcMultiBind links a provider's identity to the signed-in user. Pulled
// out so OIDCMultiAuth's branch stays one line.
func oidcMultiBind(c *gin.Context, provider *system_setting.OIDCProvider, discovery *oidc.Discovery) {
	userInfo, err := exchangeOIDCMultiUserInfo(provider, discovery, c.Query("code"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	namespacedId := provider.Name + ":" + userInfo.OpenID
	if model.IsOidcIdAlreadyTaken(namespacedId) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该 OIDC 身份已被其他账户绑定",
		})
		return
	}
	session := sessions.Default(c)
	id := session.Get("id")
	user := model.User{Id: id.(int)}
	if err := user.FillUserById(); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	user.OidcId = namespacedId
	if err := user.Update(false); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "OIDC 绑定成功"})
}

// exchangeOIDCMultiUserInfo runs the auth-code → token → userinfo dance.
// Different from the legacy single-provider helper in oidc.go because the
// endpoints come from discovery, not flat config.
func exchangeOIDCMultiUserInfo(provider *system_setting.OIDCProvider, discovery *oidc.Discovery, code string) (*OidcUser, error) {
	if code == "" {
		return nil, errors.New("无效的参数")
	}
	values := map[string]string{
		"client_id":     provider.ClientId,
		"client_secret": provider.ClientSecret,
		"code":          code,
		"grant_type":    "authorization_code",
		"redirect_uri":  fmt.Sprintf("%s/oauth/oidc/%s", config.ServerAddress, provider.Name),
	}
	jsonData, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", discovery.TokenEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := oidcMultiClient.Do(req)
	if err != nil {
		logger.SysLog("oidc multi token exchange: " + err.Error())
		return nil, errors.New("无法连接 OIDC 验证服务器，请稍后重试")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC token endpoint returned %d", res.StatusCode)
	}

	var tokenResponse OidcResponse
	if err := json.NewDecoder(res.Body).Decode(&tokenResponse); err != nil {
		return nil, err
	}

	uReq, err := http.NewRequest("GET", discovery.UserinfoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	uReq.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)
	uReq.Header.Set("Accept", "application/json")

	uRes, err := oidcMultiClient.Do(uReq)
	if err != nil {
		logger.SysLog("oidc multi userinfo fetch: " + err.Error())
		return nil, errors.New("无法获取用户信息，请稍后重试")
	}
	defer uRes.Body.Close()
	if uRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC userinfo endpoint returned %d", uRes.StatusCode)
	}

	var ou OidcUser
	if err := json.NewDecoder(uRes.Body).Decode(&ou); err != nil {
		return nil, err
	}
	if ou.OpenID == "" {
		return nil, errors.New("OIDC userinfo 缺少 sub 字段")
	}
	return &ou, nil
}
