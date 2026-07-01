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

func TestRegister_RegisterDisabled_HasCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prev := config.RegisterEnabled
	config.RegisterEnabled = false
	defer func() { config.RegisterEnabled = prev }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/user/register",
		strings.NewReader(`{"username":"x","password":"y"}`))

	Register(c)

	var body struct {
		Success bool   `json:"success"`
		Code    string `json:"code"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body.Success || body.Code != errcode.AuthRegisterDisabled {
		t.Fatalf("got success=%v code=%q, want false / %q", body.Success, body.Code, errcode.AuthRegisterDisabled)
	}
}
