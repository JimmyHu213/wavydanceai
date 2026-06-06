package oidc

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

// mockIdP serves an OIDC discovery doc and counts requests so we can prove
// the cache short-circuits the second call.
func mockIdP(t *testing.T, body string, requestCounter *int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(requestCounter, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
}

const validDiscovery = `{
  "issuer": "https://idp.test",
  "authorization_endpoint": "https://idp.test/authorize",
  "token_endpoint": "https://idp.test/token",
  "userinfo_endpoint": "https://idp.test/userinfo"
}`

func TestFetch_HappyPath(t *testing.T) {
	ClearCache()
	var count int32
	srv := mockIdP(t, validDiscovery, &count)
	defer srv.Close()

	d, err := Fetch(srv.URL)
	require.NoError(t, err)
	require.Equal(t, "https://idp.test/token", d.TokenEndpoint)
	require.Equal(t, "https://idp.test/userinfo", d.UserinfoEndpoint)
	require.EqualValues(t, 1, atomic.LoadInt32(&count))
}

func TestFetch_CachesByURL(t *testing.T) {
	ClearCache()
	var count int32
	srv := mockIdP(t, validDiscovery, &count)
	defer srv.Close()

	for i := 0; i < 5; i++ {
		_, err := Fetch(srv.URL)
		require.NoError(t, err)
	}
	require.EqualValues(t, 1, atomic.LoadInt32(&count), "second-onwards must hit cache")
}

func TestFetch_EmptyURLRejected(t *testing.T) {
	ClearCache()
	_, err := Fetch("")
	require.Error(t, err)
	_, err = Fetch("   ")
	require.Error(t, err)
}

func TestFetch_NonOKStatusRejected(t *testing.T) {
	ClearCache()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := Fetch(srv.URL)
	require.Error(t, err)
}

// An IdP returning JSON without the three endpoint fields is broken —
// reject early so the auth handler never tries to POST to "".
func TestFetch_MissingEndpointsRejected(t *testing.T) {
	ClearCache()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"issuer":"x"}`))
	}))
	defer srv.Close()

	_, err := Fetch(srv.URL)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required endpoints")
}

func TestFetch_GarbageBodyRejected(t *testing.T) {
	ClearCache()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	_, err := Fetch(srv.URL)
	require.Error(t, err)
}
