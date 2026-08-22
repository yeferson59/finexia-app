package config

import (
	"fmt"
	"strings"
	"testing"
)

// goodSecret stands in for the output of `openssl rand -base64 48`.
const goodSecret = "kP4vN8xQ2mR7wL5tZ9bC3jH6yF1sD0aG8eU4iO7pT2nV5rX"

// TestValidateRejectsUnsafeJWTSecrets is the regression guard for the fix that
// removed JWT_SECRET's default. The variable used to fall back to "secret", so
// a deployment that never set it still booted and served traffic — signing
// every access token with a value published in this repository, which is a
// forged token for any user id and any role, "admin" included.
func TestValidateRejectsUnsafeJWTSecrets(t *testing.T) {
	cases := []struct {
		name   string
		secret string
	}{
		{"unset", ""},
		{"whitespace only", "   "},
		{"the old default", "secret"},
		{"too short to carry 256 bits", "short-but-set"},
		{"placeholder under the length floor", "changeme"},
		// The three below all clear the 32-character floor, which is the point:
		// length is the rule a hurried operator satisfies by padding, so it
		// cannot be the only rule.
		{"placeholder padded to length", "my-jwt-secret-000000000000000000000000"},
		{"the old default repeated", "secretsecretsecretsecretsecretsecret"},
		{"one character repeated", strings.Repeat("a", 64)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := EnvConfig{JWTSecret: tc.secret, DatabaseURL: "postgres://localhost/db"}

			if err := env.Validate(); err == nil {
				t.Fatalf("Validate accepted JWT_SECRET %q", tc.secret)
			}
		})
	}
}

// TestValidateRejectsPlaceholderRegardlessOfCase covers the denylist for a
// value that is long enough to pass every other rule.
func TestValidateRejectsPlaceholderRegardlessOfCase(t *testing.T) {
	for _, secret := range []string{"your-256-bit-secret", "YOUR-256-BIT-SECRET"} {
		env := EnvConfig{JWTSecret: secret, DatabaseURL: "postgres://localhost/db"}

		err := env.Validate()
		if err == nil {
			t.Fatalf("Validate accepted the known placeholder %q", secret)
		}
		if !strings.Contains(err.Error(), "JWT_SECRET") {
			t.Errorf("Validate(%q) = %q, want the message to name JWT_SECRET", secret, err)
		}
	}
}

func TestValidateAcceptsAGeneratedSecret(t *testing.T) {
	env := EnvConfig{JWTSecret: goodSecret, DatabaseURL: "postgres://localhost/db"}

	if err := env.Validate(); err != nil {
		t.Fatalf("Validate rejected a well-formed configuration: %v", err)
	}
}

func TestValidateRequiresDatabaseURL(t *testing.T) {
	env := EnvConfig{JWTSecret: goodSecret}

	if err := env.Validate(); err == nil {
		t.Fatal("Validate accepted a configuration with no DATABASE_URL")
	}
}

// TestLoadEnvsHasNoJWTSecretDefault pins the other half of the fix: the loader
// must leave the field empty rather than substituting a value, so Validate is
// the thing that decides whether the process may start.
func TestLoadEnvsHasNoJWTSecretDefault(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	cfg, err := New()
	fmt.Println(cfg, err)

	if err == nil {
		t.Fatal(err)
	}

	if err.Error() != "field DATABASE_URL is required must have a value" {
		t.Fatalf("JWT_SECRET %q must not be empty", cfg.JWTSecret)
	}
}
