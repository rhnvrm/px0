package main

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Opt-in only: normal px0 keeps its original CSP and localhost POST policy.
var embedding struct{ parentOrigin, publicOrigin, token string }

func exactHTTPOrigin(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != "" && u.User == nil && u.Path == "" && u.RawQuery == "" && u.Fragment == "" && !strings.ContainsAny(raw, " \t\r\n;,'\"") && u.String() == u.Scheme+"://"+u.Host
}

func configureEmbedding(parent, public, token string) error {
	if parent == "" && public == "" {
		return nil
	}
	if !exactHTTPOrigin(parent) || !exactHTTPOrigin(public) {
		return fmt.Errorf("embed-origin and public-origin must both be exact HTTP(S) origins without paths")
	}
	if len(token) < 32 {
		return fmt.Errorf("authenticated embedding requires PX0_ACCESS_TOKEN of at least 32 characters")
	}
	embedding.parentOrigin, embedding.publicOrigin, embedding.token = parent, public, token
	return nil
}

func embeddingCSP() string {
	if embedding.parentOrigin == "" {
		return contentSecurityPolicy
	}
	return strings.Replace(contentSecurityPolicy, "frame-ancestors 'none'", "frame-ancestors "+embedding.parentOrigin, 1)
}

func embeddingPublicRequest(r *http.Request) bool {
	if embedding.publicOrigin == "" {
		return false
	}
	u, _ := url.Parse(embedding.publicOrigin)
	return r.Host == u.Host && r.Header.Get("Origin") == embedding.publicOrigin
}

func embeddingAccess(w http.ResponseWriter, r *http.Request) bool {
	if embedding.token == "" {
		return true
	}
	matches := func(t string) bool { return subtle.ConstantTimeCompare([]byte(t), []byte(embedding.token)) == 1 }
	// Bootstrap a session cookie, then remove the secret from the address bar.
	if token := r.URL.Query().Get("px0_token"); token != "" {
		if !matches(token) || r.Method != http.MethodGet {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return false
		}
		secure := strings.HasPrefix(embedding.publicOrigin, "https://")
		sameSite := http.SameSiteLaxMode
		if secure {
			sameSite = http.SameSiteNoneMode
		}
		http.SetCookie(w, &http.Cookie{Name: "px0_embed", Value: embedding.token, Path: "/", HttpOnly: true, Secure: secure, SameSite: sameSite, Partitioned: secure})
		q := r.URL.Query()
		q.Del("px0_token")
		next := *r.URL
		next.RawQuery = q.Encode()
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		http.Redirect(w, r, next.RequestURI(), http.StatusSeeOther)
		return false
	}
	if matches(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")) {
		return true
	}
	if c, err := r.Cookie("px0_embed"); err == nil && matches(c.Value) {
		return true
	}
	http.Error(w, "unauthorized", http.StatusUnauthorized)
	return false
}
