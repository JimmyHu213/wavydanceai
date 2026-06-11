// End-to-end test of the seedance adaptor through the real framework path:
// controller submit (mock Ark upstream) → task row persisted → polling round
// flips it to SUCCESS → token-based settlement charges the delta.
//
// External test package on purpose: the seedance package imports relay/task,
// so an in-package test file could not import it back.
package task_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/controller"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/channeltype"
	relaytask "github.com/songquanpeng/one-api/relay/task"
	"github.com/songquanpeng/one-api/relay/task/seedance"
)

func TestSeedanceEndToEnd(t *testing.T) {
	gin.SetMode(gin.TestMode)
	common.RedisEnabled = false
	client.Init()

	dsn := filepath.Join(t.TempDir(), "seedance-e2e.db") + "?_busy_timeout=5000"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&model.User{}, &model.Token{}, &model.Channel{}, &model.Log{}, &model.Task{},
	))
	model.DB = db
	model.LOG_DB = db

	// --- mock Ark upstream -------------------------------------------------
	const upstreamId = "cgt-20260611-e2e"
	var submitBody map[string]any
	var gotSubmitAuth, gotFetchAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v3/contents/generations/tasks":
			gotSubmitAuth = r.Header.Get("Authorization")
			require.NoError(t, json.NewDecoder(r.Body).Decode(&submitBody))
			_, _ = w.Write([]byte(`{"id":"` + upstreamId + `"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/contents/generations/tasks/"+upstreamId:
			gotFetchAuth = r.Header.Get("Authorization")
			_, _ = w.Write([]byte(`{"id":"` + upstreamId + `","status":"succeeded",` +
				`"content":{"video_url":"https://ark-content.example.com/` + upstreamId + `.mp4"},` +
				`"usage":{"completion_tokens":108900,"total_tokens":108900}}`))
		default:
			t.Errorf("unexpected upstream call: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// --- fixtures ----------------------------------------------------------
	const initialQuota = int64(1_000_000)
	suffix := random.GetUUID()[:8]
	user := &model.User{
		Username:    "sd-" + suffix,
		Password:    "x",
		Role:        model.RoleCommonUser,
		Status:      model.UserStatusEnabled,
		AccessToken: "at-" + suffix,
		AffCode:     "af-" + suffix,
		Quota:       initialQuota,
	}
	require.NoError(t, model.DB.Create(user).Error)
	token := &model.Token{
		UserId:         user.Id,
		Key:            "key-" + suffix,
		Status:         model.TokenStatusEnabled,
		Name:           "seedance-e2e",
		UnlimitedQuota: true,
	}
	require.NoError(t, model.DB.Create(token).Error)
	baseURL := server.URL
	channel := &model.Channel{
		Name:    "ark-mock",
		Type:    channeltype.Doubao,
		Key:     "sk-ark-test",
		Status:  model.ChannelStatusEnabled,
		BaseURL: &baseURL,
		Models:  strings.Join(seedance.ModelList, ","),
	}
	require.NoError(t, model.DB.Create(channel).Error)

	// --- submit through the controller --------------------------------------
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/videos", strings.NewReader(
		`{"model":"seedance-2.0","prompt":"a corgi running on the beach",`+
			`"seconds":"5","resolution":"720p","ratio":"16:9"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	// the Distribute middleware rewrites Authorization to the channel key
	c.Request.Header.Set("Authorization", "Bearer "+channel.Key)
	c.Set(ctxkey.Id, user.Id)
	c.Set(ctxkey.TokenId, token.Id)
	c.Set(ctxkey.TokenName, token.Name)
	c.Set(ctxkey.Group, "default")
	c.Set(ctxkey.Channel, channel.Type)
	c.Set(ctxkey.ChannelId, channel.Id)
	c.Set(ctxkey.RequestModel, "seedance-2.0")
	c.Set(ctxkey.BaseURL, baseURL)
	controller.RelayVideoSubmit(c)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	// upstream got an Ark-shaped create request with our key
	require.Equal(t, "Bearer sk-ark-test", gotSubmitAuth)
	require.Equal(t, "seedance-2.0", submitBody["model"])
	require.Equal(t, "720p", submitBody["resolution"])
	require.Equal(t, "16:9", submitBody["ratio"])
	require.EqualValues(t, 5, submitBody["duration"])

	// the client got an OpenAI-Video-shaped queued task with a local id
	var videoResp struct {
		Id     string `json:"id"`
		Object string `json:"object"`
		Status string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &videoResp))
	require.True(t, strings.HasPrefix(videoResp.Id, "task_"), videoResp.Id)
	require.Equal(t, "video", videoResp.Object)
	require.Equal(t, "queued", videoResp.Status)

	// task row persisted with the pre-consumed estimate:
	// 5s × 1280 × 720 × 24 / 1024 = 108000 tokens × ratio 3.5 = 378000 quota
	const estimatedQuota = int64(378_000)
	task, err := model.GetTaskByTaskId(videoResp.Id)
	require.NoError(t, err)
	require.Equal(t, seedance.Platform, task.Platform)
	require.Equal(t, model.TaskStatusQueued, task.Status)
	require.Equal(t, estimatedQuota, task.Quota)
	pd, err := task.GetPrivateData()
	require.NoError(t, err)
	require.Equal(t, upstreamId, pd.UpstreamTaskId)
	require.Equal(t, initialQuota-estimatedQuota, currentQuota(t, user.Id))

	// --- one real polling round ---------------------------------------------
	relaytask.PollOnceForTest(c.Request.Context())

	require.Equal(t, "Bearer sk-ark-test", gotFetchAuth)
	task, err = model.GetTaskByTaskId(videoResp.Id)
	require.NoError(t, err)
	require.Equal(t, model.TaskStatusSuccess, task.Status)
	require.Equal(t, 100, task.Progress)
	require.Equal(t, "https://ark-content.example.com/"+upstreamId+".mp4", task.ResultUrl)

	// settlement recalculated from upstream usage: 108900 × 3.5 = 381150,
	// i.e. a 3150 delta charged on top of the estimate
	const settledQuota = int64(381_150)
	require.Equal(t, settledQuota, task.Quota)
	require.Equal(t, initialQuota-settledQuota, currentQuota(t, user.Id))

	var consumeLogs int64
	require.NoError(t, model.LOG_DB.Model(&model.Log{}).
		Where("type = ?", model.LogTypeConsume).Count(&consumeLogs).Error)
	require.EqualValues(t, 2, consumeLogs, "submit charge + settlement delta")
	var refundLogs int64
	require.NoError(t, model.LOG_DB.Model(&model.Log{}).
		Where("type = ?", model.LogTypeRefund).Count(&refundLogs).Error)
	require.Zero(t, refundLogs)
}

func currentQuota(t *testing.T, userId int) int64 {
	t.Helper()
	var user model.User
	require.NoError(t, model.DB.First(&user, userId).Error)
	return user.Quota
}
