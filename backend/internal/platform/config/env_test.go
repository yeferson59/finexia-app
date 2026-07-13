package config

import (
	"testing"
	"time"
)

func TestGetString(t *testing.T) {
	c := New()

	t.Run("returns default when unset", func(t *testing.T) {
		if got := c.getString("TEST_UNSET_STRING", "fallback"); got != "fallback" {
			t.Errorf("got %q, want fallback", got)
		}
	})

	t.Run("returns value and trims spaces", func(t *testing.T) {
		t.Setenv("TEST_SET_STRING", "  value  ")
		if got := c.getString("TEST_SET_STRING", "fallback"); got != "value" {
			t.Errorf("got %q, want value", got)
		}
	})

	t.Run("blank value falls back to default", func(t *testing.T) {
		t.Setenv("TEST_BLANK_STRING", "   ")
		if got := c.getString("TEST_BLANK_STRING", "fallback"); got != "fallback" {
			t.Errorf("got %q, want fallback", got)
		}
	})
}

func TestGetDuration(t *testing.T) {
	c := New()
	def := 15 * time.Minute

	cases := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{"seconds", "45s", 45 * time.Second},
		{"minutes", "10m", 10 * time.Minute},
		{"hours", "2h", 2 * time.Hour},
		{"days", "7d", 7 * 24 * time.Hour},
		{"invalid number", "xxs", def},
		{"no unit", "30", def},
		{"garbage", "soon", def},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TEST_DURATION", tc.value)
			if got := c.getDuration("TEST_DURATION", def); got != tc.want {
				t.Errorf("getDuration(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}

	t.Run("unset returns default", func(t *testing.T) {
		if got := c.getDuration("TEST_UNSET_DURATION", def); got != def {
			t.Errorf("got %v, want %v", got, def)
		}
	})
}

func TestGetBool(t *testing.T) {
	c := New()

	cases := []struct {
		name  string
		value string
		def   bool
		want  bool
	}{
		{"true", "true", false, true},
		{"false", "false", true, false},
		{"numeric one", "1", false, true},
		{"invalid keeps default", "not-a-bool", true, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TEST_BOOL", tc.value)
			if got := c.getBool("TEST_BOOL", tc.def); got != tc.want {
				t.Errorf("getBool(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}

	t.Run("unset returns default", func(t *testing.T) {
		if got := c.getBool("TEST_UNSET_BOOL", true); got != true {
			t.Error("expected default true")
		}
	})
}

func TestGetSlice(t *testing.T) {
	c := New()

	t.Run("unset returns defaults", func(t *testing.T) {
		got := c.getSlice("TEST_UNSET_SLICE", "a", "b")
		if len(got) != 2 || got[0] != "a" || got[1] != "b" {
			t.Errorf("got %v, want [a b]", got)
		}
	})

	t.Run("single value", func(t *testing.T) {
		t.Setenv("TEST_SLICE", "one")
		got := c.getSlice("TEST_SLICE", "default")
		if len(got) != 1 || got[0] != "one" {
			t.Errorf("got %v, want [one]", got)
		}
	})

	t.Run("comma separated values", func(t *testing.T) {
		t.Setenv("TEST_SLICE", "one,two,three")
		got := c.getSlice("TEST_SLICE", "default")
		if len(got) != 3 || got[0] != "one" || got[1] != "two" || got[2] != "three" {
			t.Errorf("got %v, want [one two three]", got)
		}
	})
}

func TestLoadEnvs(t *testing.T) {
	c := New()

	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("PORT", "9090")
	t.Setenv("JWT_SECRET", "super-secret")
	t.Setenv("JWT_ACCESS_DURATION", "30m")
	t.Setenv("JWT_REFRESH_DURATION", "14d")
	t.Setenv("CORS_ENABLED", "false")
	t.Setenv("CORS_ORIGIN", "https://a.example,https://b.example")

	env := c.LoadEnvs()

	if env.Environment != "production" {
		t.Errorf("Environment = %q, want production", env.Environment)
	}
	if env.Port != "9090" {
		t.Errorf("Port = %q, want 9090", env.Port)
	}
	if env.JWTSecret != "super-secret" {
		t.Errorf("JWTSecret = %q, want super-secret", env.JWTSecret)
	}
	if env.JWTAccessDuration != 30*time.Minute {
		t.Errorf("JWTAccessDuration = %v, want 30m", env.JWTAccessDuration)
	}
	if env.JWTRefreshDuration != 14*24*time.Hour {
		t.Errorf("JWTRefreshDuration = %v, want 14d", env.JWTRefreshDuration)
	}
	if env.CORSEnabled {
		t.Error("CORSEnabled = true, want false")
	}
	if len(env.CORSOrigin) != 2 || env.CORSOrigin[0] != "https://a.example" {
		t.Errorf("CORSOrigin = %v, want the two configured origins", env.CORSOrigin)
	}
}
