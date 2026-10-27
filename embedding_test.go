package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmbeddingConfiguration(t *testing.T) {
	old := embedding
	defer func() { embedding = old }()
	for _, origin := range []string{"*", "https://bb.example/", "https://bb.example; script-src *", "javascript:alert(1)", "https://user@bb.example", "https://bb.example?x=1"} {
		if configureEmbedding(origin, "https://px.example", strings.Repeat("x", 32)) == nil {
			t.Fatalf("accepted %q", origin)
		}
	}
	if configureEmbedding("https://bb.example", "https://px.example", "short") == nil {
		t.Fatal("accepted short token")
	}
	if err := configureEmbedding("https://bb.example", "https://px.example", strings.Repeat("x", 32)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(embeddingCSP(), "frame-ancestors https://bb.example;") {
		t.Fatal(embeddingCSP())
	}
	r := httptest.NewRequest("POST", "https://px.example/api/file/save", nil)
	r.Header.Set("Origin", "https://px.example")
	if !localPost(httptest.NewRecorder(), r) {
		t.Fatal("explicit proxy origin rejected")
	}
	r.Header.Set("Origin", "https://evil.example")
	if localPost(httptest.NewRecorder(), r) {
		t.Fatal("cross-origin write accepted")
	}
	r.Host = "other.example"
	r.Header.Set("Origin", "https://other.example")
	if localPost(httptest.NewRecorder(), r) {
		t.Fatal("unconfigured hostname accepted")
	}
}
func TestEmbeddingAuthentication(t *testing.T) {
	old := embedding
	defer func() { embedding = old }()
	token := strings.Repeat("a", 64)
	if err := configureEmbedding("https://bb.example", "https://px.example", token); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/", "/api/meta", "/static/app.js"} {
		w := httptest.NewRecorder()
		if embeddingAccess(w, httptest.NewRequest("GET", path, nil)) || w.Code != 401 {
			t.Fatal("unauthenticated access", path)
		}
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/?path=hello.go&px0_token="+token, nil)
	if embeddingAccess(w, r) || w.Code != http.StatusSeeOther {
		t.Fatal("bootstrap failed")
	}
	if w.Header().Get("Location") != "/?path=hello.go" {
		t.Fatal("token not stripped")
	}
	c := w.Result().Cookies()[0]
	if !c.HttpOnly || !c.Secure || !c.Partitioned {
		t.Fatal("insecure cookie")
	}
	r = httptest.NewRequest("GET", "/api/meta", nil)
	r.AddCookie(c)
	if !embeddingAccess(httptest.NewRecorder(), r) {
		t.Fatal("cookie rejected")
	}
	r = httptest.NewRequest("GET", "/api/meta", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	if !embeddingAccess(httptest.NewRecorder(), r) {
		t.Fatal("bearer rejected")
	}
}
