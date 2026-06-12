package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/i18n"
	"github.com/songquanpeng/one-api/model"

	"github.com/gin-gonic/gin"
)

func GetOptions(c *gin.Context) {
	var options []*model.Option
	config.OptionMapRWMutex.Lock()
	for k, v := range config.OptionMap {
		if strings.HasSuffix(k, "Token") || strings.HasSuffix(k, "Secret") {
			continue
		}
		options = append(options, &model.Option{
			Key:   k,
			Value: helper.Interface2String(v),
		})
	}
	config.OptionMapRWMutex.Unlock()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    options,
	})
	return
}

func UpdateOption(c *gin.Context) {
	var option model.Option
	err := json.NewDecoder(c.Request.Body).Decode(&option)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	if msg := optionUpdateRejection(option.Key, option.Value); msg != "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": msg,
		})
		return
	}
	err = model.UpdateOption(option.Key, option.Value)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

// optionUpdateRejection returns a non-empty rejection message if setting key to
// value is not allowed (e.g. enabling an OAuth provider before its credentials
// are configured), or "" if the update may proceed. Shared by the single-key
// UpdateOption and the batch UpdateOptionsBatch so both validate identically.
func optionUpdateRejection(key, value string) string {
	switch key {
	case "Theme":
		if !config.ValidThemes[value] {
			return "无效的主题"
		}
	case "GitHubOAuthEnabled":
		if value == "true" && config.GitHubClientId == "" {
			return "无法启用 GitHub OAuth，请先填入 GitHub Client Id 以及 GitHub Client Secret！"
		}
	case "EmailDomainRestrictionEnabled":
		if value == "true" && len(config.EmailDomainWhitelist) == 0 {
			return "无法启用邮箱域名限制，请先填入限制的邮箱域名！"
		}
	case "WeChatAuthEnabled":
		if value == "true" && config.WeChatServerAddress == "" {
			return "无法启用微信登录，请先填入微信登录相关配置信息！"
		}
	case "TurnstileCheckEnabled":
		if value == "true" && config.TurnstileSiteKey == "" {
			return "无法启用 Turnstile 校验，请先填入 Turnstile 校验相关配置信息！"
		}
	}
	return ""
}

// UpdateOptionsBatch persists several options atomically — either all the
// supplied keys are written or none is. Callers like the pricing editor use it
// so two related ratio maps (ModelRatio + CompletionRatio) can never end up
// half-saved.
func UpdateOptionsBatch(c *gin.Context) {
	var req struct {
		Keys map[string]string `json:"keys"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil || len(req.Keys) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	for key, value := range req.Keys {
		if msg := optionUpdateRejection(key, value); msg != "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": msg,
			})
			return
		}
	}
	if err := model.UpdateOptions(req.Keys); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}
