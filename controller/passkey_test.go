package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	vwa "github.com/descope/virtualwebauthn"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/model"
	settingpkg "github.com/songquanpeng/one-api/setting/passkey"
)

// setupPasskeyCtrlTest gives each test its own in-memory DB, gin engine
// with cookie sessions, and a valid passkey setting tuned for localhost.
// Returns the engine + the seeded user + a virtual authenticator + RP.
func setupPasskeyCtrlTest(t *testing.T) (*gin.Engine, *model.User, vwa.Authenticator, vwa.Credential, vwa.RelyingParty) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:?_busy_timeout=5000"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.PasskeyCredential{}, &model.Log{}, &model.Option{}))
	model.DB = db
	model.LOG_DB = db

	s := settingpkg.GetPasskeySetting()
	prev := *s
	t.Cleanup(func() { *s = prev })
	s.Enabled = true
	s.RPID = "localhost"
	s.RPName = "wavydance"
	s.RPOrigins = `["http://localhost"]`

	u := &model.User{Username: "alice-c", Password: "x", Status: model.UserStatusEnabled, Role: 1}
	require.NoError(t, db.Create(u).Error)

	engine := gin.New()
	engine.Use(sessions.Sessions("wavy", cookie.NewStore([]byte("test-secret"))))
	stubAuth := func(c *gin.Context) {
		c.Set(ctxkey.Id, u.Id)
		c.Next()
	}
	engine.GET("/api/user/passkey/credentials", stubAuth, ListMyPasskeys)
	engine.POST("/api/user/passkey/credentials/register/begin", stubAuth, BeginRegisterPasskey)
	engine.POST("/api/user/passkey/credentials/register/finish", stubAuth, FinishRegisterPasskey)
	engine.PATCH("/api/user/passkey/credentials/:id", stubAuth, RenameMyPasskey)
	engine.DELETE("/api/user/passkey/credentials/:id", stubAuth, DeleteMyPasskey)

	rp := vwa.RelyingParty{Name: "wavydance", ID: "localhost", Origin: "http://localhost"}
	return engine, u, vwa.NewAuthenticator(), vwa.NewCredential(vwa.KeyTypeEC2), rp
}

func TestListMyPasskeys_Empty(t *testing.T) {
	engine, _, _, _, _ := setupPasskeyCtrlTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/user/passkey/credentials", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.True(t, body["success"].(bool))
	require.Empty(t, body["data"])
}

func TestRegisterRoundTrip(t *testing.T) {
	engine, _, auth, cred, rp := setupPasskeyCtrlTest(t)

	beginBody, _ := json.Marshal(map[string]string{"name": "MacBook"})
	beginReq := httptest.NewRequest(http.MethodPost, "/api/user/passkey/credentials/register/begin", bytes.NewReader(beginBody))
	beginReq.Header.Set("Content-Type", "application/json")
	beginRec := httptest.NewRecorder()
	engine.ServeHTTP(beginRec, beginReq)
	require.Equal(t, http.StatusOK, beginRec.Code, beginRec.Body.String())

	setCookie := beginRec.Result().Header.Get("Set-Cookie")
	require.NotEmpty(t, setCookie, "begin must set a session cookie carrying the challenge")

	var beginEnv struct {
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(beginRec.Body.Bytes(), &beginEnv))

	// API: ParseAttestationOptions takes a string and returns (*AttestationOptions, error)
	attestationOptions, err := vwa.ParseAttestationOptions(string(beginEnv.Data))
	require.NoError(t, err)
	attestation := vwa.CreateAttestationResponse(rp, auth, cred, *attestationOptions)

	finishReq := httptest.NewRequest(http.MethodPost, "/api/user/passkey/credentials/register/finish", bytes.NewReader([]byte(attestation)))
	finishReq.Header.Set("Content-Type", "application/json")
	finishReq.Header.Set("Cookie", setCookie)
	finishRec := httptest.NewRecorder()
	engine.ServeHTTP(finishRec, finishReq)
	require.Equal(t, http.StatusOK, finishRec.Code, finishRec.Body.String())

	creds, err := model.ListPasskeysByUserId(1)
	require.NoError(t, err)
	require.Len(t, creds, 1)
	require.Equal(t, "MacBook", creds[0].Name)
}

func TestFinishRegisterWithoutPendingChallenge(t *testing.T) {
	engine, _, _, _, _ := setupPasskeyCtrlTest(t)
	req := httptest.NewRequest(http.MethodPost, "/api/user/passkey/credentials/register/finish", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestRenameAndDelete(t *testing.T) {
	engine, u, _, _, _ := setupPasskeyCtrlTest(t)
	cred := &model.PasskeyCredential{UserId: u.Id, CredentialId: []byte{1, 2, 3}, PublicKey: []byte{9}, Name: "old", CreatedAt: time.Now().Unix()}
	require.NoError(t, model.CreatePasskey(cred))

	renameBody, _ := json.Marshal(map[string]string{"name": "new"})
	req := httptest.NewRequest(http.MethodPatch, "/api/user/passkey/credentials/1", bytes.NewReader(renameBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	got, err := model.GetPasskeyByIdForUser(cred.Id, u.Id)
	require.NoError(t, err)
	require.Equal(t, "new", got.Name)

	delReq := httptest.NewRequest(http.MethodDelete, "/api/user/passkey/credentials/1", nil)
	delW := httptest.NewRecorder()
	engine.ServeHTTP(delW, delReq)
	require.Equal(t, http.StatusOK, delW.Code)

	all, _ := model.ListPasskeysByUserId(u.Id)
	require.Empty(t, all)
}

func TestDeleteForeignReturns404(t *testing.T) {
	engine, _, _, _, _ := setupPasskeyCtrlTest(t)
	other := &model.User{Username: "bob", Password: "x", Status: model.UserStatusEnabled, Role: 1, AccessToken: "bob-tok-1234567890123456", AffCode: "b0b1"}
	require.NoError(t, model.DB.Create(other).Error)
	require.NoError(t, model.CreatePasskey(&model.PasskeyCredential{UserId: other.Id, CredentialId: []byte{0xff}, PublicKey: []byte{1}, CreatedAt: time.Now().Unix()}))

	req := httptest.NewRequest(http.MethodDelete, "/api/user/passkey/credentials/1", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestPasskeyDisabledRejects403(t *testing.T) {
	engine, _, _, _, _ := setupPasskeyCtrlTest(t)
	settingpkg.GetPasskeySetting().Enabled = false
	req := httptest.NewRequest(http.MethodGet, "/api/user/passkey/credentials", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	req2 := httptest.NewRequest(http.MethodPost, "/api/user/passkey/credentials/register/begin", bytes.NewReader([]byte(`{}`)))
	w2 := httptest.NewRecorder()
	engine.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusForbidden, w2.Code)
}
