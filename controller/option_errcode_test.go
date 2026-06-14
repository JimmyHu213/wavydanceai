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

// Setting Theme to an unknown value is rejected by optionUpdateRejection
// before any DB write, so this asserts the coded option.invalid_value path
// without a DB.
func TestUpdateOption_InvalidTheme_HasCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/option",
		strings.NewReader(`{"key":"Theme","value":"__not_a_real_theme__"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateOption(c)

	var body struct {
		Success bool   `json:"success"`
		Code    string `json:"code"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body.Success || body.Code != errcode.OptionInvalidValue {
		t.Fatalf("got success=%v code=%q, want false / %q", body.Success, body.Code, errcode.OptionInvalidValue)
	}
}
