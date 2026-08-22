package config

import (
	"testing"
	"time"
)

func TestLoadEnvs(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("PORT", "9090")
	t.Setenv("JWT_SECRET", "super-secret")
	t.Setenv("JWT_ACCESS_DURATION", "30m")
	t.Setenv("JWT_REFRESH_DURATION", "14d")
	t.Setenv("CORS_ENABLED", "false")
	t.Setenv("CORS_ORIGIN", "https://a.example,https://b.example")

	c, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if c.Environment != "production" {
		t.Errorf("Environment = %q, want production", c.Environment)
	}
	if c.Port != "9090" {
		t.Errorf("Port = %q, want 9090", c.Port)
	}
	if c.JWTSecret != "super-secret" {
		t.Errorf("JWTSecret = %q, want super-secret", c.JWTSecret)
	}
	if c.JWTAccessDuration != 30*time.Minute {
		t.Errorf("JWTAccessDuration = %v, want 30m", c.JWTAccessDuration)
	}
	if c.JWTRefreshDuration != 14*24*time.Hour {
		t.Errorf("JWTRefreshDuration = %v, want 14d", c.JWTRefreshDuration)
	}
	if c.CORSEnabled {
		t.Error("CORSEnabled = true, want false")
	}

	cors := c.CORSOrigin

	if len(cors) != 2 || cors[0] != "https://a.example" {
		t.Errorf("CORSOrigin = %v, want the two configured origins", c.CORSOrigin)
	}
}
