package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/model"
)

func setupOptionCtrlTest(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:?_busy_timeout=5000"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Option{}))
	model.DB = db
	config.OptionMap = make(map[string]string)

	e := gin.New()
	e.PUT("/option/batch", UpdateOptionsBatch)
	return e
}

func putBatch(e *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, "/option/batch", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	return w
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &out))
	return out
}

func TestUpdateOptionsBatchPersistsAtomically(t *testing.T) {
	e := setupOptionCtrlTest(t)
	w := putBatch(e, `{"keys":{"ModelRatio":"{\"gpt-4o\":1.5}","CompletionRatio":"{\"gpt-4o\":4}"}}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, true, decodeJSON(t, w)["success"])
	require.Equal(t, `{"gpt-4o":1.5}`, config.OptionMap["ModelRatio"])
	require.Equal(t, `{"gpt-4o":4}`, config.OptionMap["CompletionRatio"])
}

func TestUpdateOptionsBatchRejectsInvalidGuard(t *testing.T) {
	e := setupOptionCtrlTest(t)
	config.GitHubClientId = "" // enabling GitHub OAuth without a client id is rejected
	w := putBatch(e, `{"keys":{"GitHubOAuthEnabled":"true"}}`)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, false, decodeJSON(t, w)["success"])
	_, has := config.OptionMap["GitHubOAuthEnabled"]
	require.False(t, has, "a rejected batch must persist nothing")
}

func TestUpdateOptionsBatchRejectsEmptyBody(t *testing.T) {
	e := setupOptionCtrlTest(t)
	require.Equal(t, http.StatusBadRequest, putBatch(e, `{"keys":{}}`).Code)
	require.Equal(t, http.StatusBadRequest, putBatch(e, `not json`).Code)
}
