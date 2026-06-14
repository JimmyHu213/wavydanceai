package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/errcode"
)

func TestSendError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	SendError(c, http.StatusOK, errcode.ParamInvalid, "参数错误")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body struct {
		Success bool   `json:"success"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Success {
		t.Error("success = true, want false")
	}
	if body.Code != errcode.ParamInvalid {
		t.Errorf("code = %q, want %q", body.Code, errcode.ParamInvalid)
	}
	if body.Message != "参数错误" {
		t.Errorf("message = %q, want 参数错误", body.Message)
	}
}
