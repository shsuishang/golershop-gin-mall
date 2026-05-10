package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golershop.cn/internal/config"
)

func TestLoginResponseShape(t *testing.T) {
	cfg := &config.AppConfig{
		JWT: config.JWTConfig{
			TokenHeader: "Authorization",
			TokenPrefix: "Bearer ",
			TokenSecret: "test-secret",
		},
		Secure: config.SecureConfig{
			Ignore: []string{"/front/account/login/login"},
		},
	}
	r := New(cfg)

	body := strings.NewReader(`{"userAccount":"admin","password":"123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/front/account/login/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected http code: %d", w.Code)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	for _, key := range []string{"status", "code", "msg", "data"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("missing key: %s", key)
		}
	}
}

func TestAuthRequired(t *testing.T) {
	cfg := &config.AppConfig{
		JWT: config.JWTConfig{
			TokenHeader: "Authorization",
			TokenPrefix: "Bearer ",
			TokenSecret: "test-secret",
		},
		Secure: config.SecureConfig{
			Ignore: []string{"/front/account/login/login"},
		},
	}
	r := New(cfg)

	req := httptest.NewRequest(http.MethodGet, "/front/trade/order/list", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var m map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &m)
	if m["msg"] != "需要登录" {
		t.Fatalf("unexpected msg: %v", m["msg"])
	}
}
