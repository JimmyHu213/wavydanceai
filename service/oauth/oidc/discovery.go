// Package oidc handles OpenID Connect discovery — fetching and caching the
// well-known/openid-configuration document so the auth handler doesn't
// re-hit the IdP every time someone signs in.
//
// Cache is process-local and never invalidates. IdPs rotate endpoints
// extremely rarely (and a deploy/restart picks up changes). Adding an
// invalidation channel would be over-engineering for the size of this
// problem.
package oidc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Discovery is the subset of the OIDC discovery document we actually use
// for auth-code flow. JWKS / supported scopes / etc. are deliberately
// dropped — adding them later means adding fields here, no other change.
type Discovery struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
}

var (
	cache     = map[string]*Discovery{}
	cacheLock sync.RWMutex
	client    = &http.Client{Timeout: 10 * time.Second}
)

// Fetch returns the discovery document for wellKnownURL, cached after the
// first successful fetch. Concurrent callers for the same URL will each do
// their own fetch — the cost is one duplicate request on first cold start,
// which beats the lock contention of a singleflight for this access rate.
func Fetch(wellKnownURL string) (*Discovery, error) {
	if strings.TrimSpace(wellKnownURL) == "" {
		return nil, errors.New("well-known URL is empty")
	}

	cacheLock.RLock()
	if d, ok := cache[wellKnownURL]; ok {
		cacheLock.RUnlock()
		return d, nil
	}
	cacheLock.RUnlock()

	res, err := client.Get(wellKnownURL)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oidc discovery: %s returned %d", wellKnownURL, res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery: read body: %w", err)
	}
	var d Discovery
	if err := json.Unmarshal(body, &d); err != nil {
		return nil, fmt.Errorf("oidc discovery: parse: %w", err)
	}
	if d.AuthorizationEndpoint == "" || d.TokenEndpoint == "" || d.UserinfoEndpoint == "" {
		return nil, fmt.Errorf("oidc discovery: %s missing required endpoints", wellKnownURL)
	}

	cacheLock.Lock()
	cache[wellKnownURL] = &d
	cacheLock.Unlock()
	return &d, nil
}

// ClearCache drops every cached entry. Currently only used by tests; if an
// admin endpoint to force re-discovery is ever added, it can call this.
func ClearCache() {
	cacheLock.Lock()
	cache = map[string]*Discovery{}
	cacheLock.Unlock()
}
