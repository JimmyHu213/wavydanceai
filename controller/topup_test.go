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

// With payments disabled, RequestTopupAmount short-circuits before any DB
// access, so this asserts the coded topup.payments_disabled path without a DB.
func TestRequestTopupAmount_PaymentsDisabled_HasCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prev := config.PaymentEnabled
	config.PaymentEnabled = false
	defer func() { config.PaymentEnabled = prev }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/user/topup/amount",
		strings.NewReader(`{"money":1000}`))
	c.Request.Header.Set("Content-Type", "application/json")

	RequestTopupAmount(c)

	var body struct {
		Success bool   `json:"success"`
		Code    string `json:"code"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body.Success || body.Code != errcode.TopupPaymentsDisabled {
		t.Fatalf("got success=%v code=%q, want false / %q", body.Success, body.Code, errcode.TopupPaymentsDisabled)
	}
}
