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
