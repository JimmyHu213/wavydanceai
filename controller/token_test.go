package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/errcode"
)

// A token name longer than 30 chars fails validateToken before any DB or
// context access, so this asserts the coded param.invalid path without a DB.
func TestAddToken_InvalidName_HasCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/token",
		strings.NewReader(`{"name":"this-token-name-is-definitely-longer-than-thirty-chars"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	AddToken(c)

	var body struct {
		Success bool   `json:"success"`
		Code    string `json:"code"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body.Success || body.Code != errcode.ParamInvalid {
		t.Fatalf("got success=%v code=%q, want false / %q", body.Success, body.Code, errcode.ParamInvalid)
	}
}
