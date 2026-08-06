package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Env struct {
	Environment        string
	Port               string
	PathMigration      string
	DatabaseURL        string
	RedisHost          string
	RedisPort          string
	RedisPassword      string
	RedisDB            int
	JWTSecret          string
	JWTAccessDuration  time.Duration
	JWTRefreshDuration time.Duration
	RefreshGracePeriod time.Duration
	MaxLoginAttempts   int
	LoginLockout       time.Duration
	TrustProxy         bool
	TrustedProxies     []string
	CORSEnabled        bool
	CORSOrigin         []string
	AWSAccessKeyID     string
	AWSDefaultRegion   string
	AWSEndpointURL     string
	AWSS3BucketName    string
	AWSSecretAccessKey string
	ResendAPIKey       string
	EmailFrom          string
	// Market data is BYO-key: each user supplies their own provider key, so
	// the application holds no provider credentials of its own. These two are
	// the keys that seal the users' keys, not keys to any provider.
	MarketKEKKeys           string
	MarketKEKActive         string
	PublicURL               string
	FrontendURL             string
	InvitationExpiry        time.Duration
	PasswordResetExpiry     time.Duration
	EmailVerificationExpiry time.Duration
	SelfRegistrationEnabled bool
	TwoFactorPendingExpiry  time.Duration
}

func (c *Config) LoadEnvs() *Env {
	_ = godotenv.Load()

	return new(Env{
		Environment:   c.getString("ENVIRONMENT", "development"),
		Port:          c.getString("PORT", "8080"),
		PathMigration: c.getString("PATH_MIGRATION", "file://internal/migrations"),
		DatabaseURL:   c.getString("DATABASE_URL", ""),
		// No default. The secret signs every access token, so a fallback value
		// would let anyone who has read this repository mint a token for any
		// user id and any role — including "admin" — against a deployment that
		// simply forgot to set the variable. Validate makes the process refuse
		// to start instead, the same treatment MARKET_KEK_KEYS already gets.
		JWTSecret:          c.getString("JWT_SECRET", ""),
		RedisHost:          c.getString("REDIS_HOST", "localhost"),
		RedisPort:          c.getString("REDIS_PORT", "6379"),
		RedisPassword:      c.getString("REDIS_PASSWORD", ""),
		RedisDB:            c.getInt("REDIS_DB", 0),
		JWTAccessDuration:  c.getDuration("JWT_ACCESS_DURATION", 15*time.Minute),
		JWTRefreshDuration: c.getDuration("JWT_REFRESH_DURATION", 30*24*time.Hour),
		RefreshGracePeriod: c.getDuration("JWT_REFRESH_GRACE_PERIOD", 30*time.Second),
		MaxLoginAttempts:   c.getInt("MAX_LOGIN_ATTEMPTS", 5),
		LoginLockout:       c.getDuration("LOGIN_LOCKOUT_DURATION", 15*time.Minute),
		TrustProxy:         c.getBool("TRUST_PROXY", true),
		TrustedProxies:     c.getSlice("TRUSTED_PROXIES"),
		CORSEnabled:        c.getBool("CORS_ENABLED", true),
		CORSOrigin:         c.getSlice("CORS_ORIGIN", "http://localhost:5173"),
		AWSAccessKeyID:     c.getString("AWS_ACCESS_KEY_ID", ""),
		AWSDefaultRegion:   c.getString("AWS_DEFAULT_REGION", ""),
		AWSEndpointURL:     c.getString("AWS_ENDPOINT_URL", ""),
		AWSS3BucketName:    c.getString("AWS_S3_BUCKET_NAME", ""),
		AWSSecretAccessKey: c.getString("AWS_SECRET_ACCESS_KEY", ""),
		ResendAPIKey:       c.getString("RESEND_API_KEY", ""),
		EmailFrom:          c.getString("EMAIL_FROM", "Finexia <noreply@finexia.me>"),
		// The key-encryption keys that wrap each user's market-data API key.
		//
		// MARKET_KEK_KEYS is a comma-separated list of "version:base64key"
		// entries, where version is a decimal integer and the key is standard
		// base64 (with padding) that must decode to exactly 32 bytes — AES-256.
		// Generate one with `openssl rand -base64 32`. Several versions may be
		// held at once so a key can be rotated without a downtime window; the
		// rows sealed under the old version stay readable while they are
		// re-wrapped.
		//
		// MARKET_KEK_ACTIVE names the version new credentials are sealed under,
		// and must be one of the versions supplied above.
		//
		// The empty default does not make this optional: it is what lets the
		// error come from secretbox, which can say which of "no key at all",
		// "not 32 bytes" and "active version not supplied" went wrong. main
		// refuses to start either way — sealing users' keys under a guessable
		// default would be worse than not starting.
		MarketKEKKeys:           c.getString("MARKET_KEK_KEYS", ""),
		MarketKEKActive:         c.getString("MARKET_KEK_ACTIVE", "1"),
		PublicURL:               c.getString("PUBLIC_URL", "http://localhost:8080"),
		FrontendURL:             c.getString("FRONTEND_URL", "http://localhost:5173"),
		InvitationExpiry:        c.getDuration("INVITATION_EXPIRY", 72*time.Hour),
		PasswordResetExpiry:     c.getDuration("PASSWORD_RESET_EXPIRY", 1*time.Hour),
		EmailVerificationExpiry: c.getDuration("EMAIL_VERIFICATION_EXPIRY", 24*time.Hour),
		// Off by default: the product is invite-only during the beta, so
		// public self-registration must be explicitly opted into.
		SelfRegistrationEnabled: c.getBool("SELF_REGISTRATION_ENABLED", false),
		// How long a password-validated login may wait for its TOTP code
		// before the user must start over.
		TwoFactorPendingExpiry: c.getDuration("TWO_FACTOR_PENDING_EXPIRY", 5*time.Minute),
	})
}

// minJWTSecretLen is the shortest JWT_SECRET the API will start with. HS256
// keys should carry at least as much entropy as the digest they feed, so 32
// bytes is the floor; `openssl rand -base64 48` produces a suitable value.
const minJWTSecretLen = 32

// minJWTSecretDistinctChars is the smallest alphabet a real key would draw
// from. It is what stops the length rule being satisfied by repetition:
// "secretsecretsecretsecretsecretsecret" is 36 characters long and has five.
const minJWTSecretDistinctChars = 8

// jwtSecretPlaceholders are matched as substrings, not compared for equality,
// because the value that actually shows up in a deployment is a placeholder
// padded until the length check stopped complaining. Every entry is a phrase
// specific enough that a random key will not contain it.
var jwtSecretPlaceholders = []string{
	"changeme",
	"change-me",
	"change_me",
	"your-256-bit-secret",
	"replace-with",
	"replace_with",
	"replacewith",
	"supersecret",
	"super-secret",
	"my-secret",
	"jwtsecret",
	"jwt-secret",
	"jwt_secret",
	"insecure",
	"placeholder",
	"example",
	"todo",
}

// Validate reports the first configuration error that must stop the process.
//
// It covers the settings whose absence is not a degraded mode but a security
// hole: a deployment missing them would run, serve traffic, and be trivially
// compromised. Everything else in Env has a defensible default and is left to
// the module that consumes it.
func (e *Env) Validate() error {
	if err := validateJWTSecret(e.JWTSecret); err != nil {
		return err
	}

	if e.DatabaseURL == "" {
		return errors.New("config: DATABASE_URL is required")
	}

	return nil
}

// validateJWTSecret refuses the signing keys that would make every access token
// forgeable: absent, too small to carry 256 bits, an obvious placeholder, or a
// short string repeated up to the length floor.
func validateJWTSecret(raw string) error {
	const generate = " — generate one with `openssl rand -base64 48`"

	secret := strings.TrimSpace(raw)

	if secret == "" {
		return errors.New("config: JWT_SECRET is required" + generate)
	}

	if len(secret) < minJWTSecretLen {
		return fmt.Errorf("config: JWT_SECRET must be at least %d characters (got %d)%s", minJWTSecretLen, len(secret), generate)
	}

	lowered := strings.ToLower(secret)
	for _, placeholder := range jwtSecretPlaceholders {
		if strings.Contains(lowered, placeholder) {
			return fmt.Errorf("config: JWT_SECRET looks like a placeholder (contains %q)%s", placeholder, generate)
		}
	}

	distinct := make(map[rune]struct{}, minJWTSecretDistinctChars)
	for _, r := range secret {
		distinct[r] = struct{}{}
	}
	if len(distinct) < minJWTSecretDistinctChars {
		return fmt.Errorf("config: JWT_SECRET uses only %d distinct characters, so its length does not reflect real entropy%s", len(distinct), generate)
	}

	return nil
}

func (Config) getString(key, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value != "" {
		return value
	}

	return defaultValue
}

func (Config) getDuration(key string, defaultValue time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value == "" {
		return defaultValue
	}

	if strings.Contains(value, "s") {
		before, _, _ := strings.Cut(value, "s")
		int64Value, err := strconv.ParseInt(before, 10, 64)
		if err != nil {
			return defaultValue
		}

		return time.Second * time.Duration(int64Value)
	}

	if strings.Contains(value, "m") {
		before, _, _ := strings.Cut(value, "m")
		int64Value, err := strconv.ParseInt(before, 10, 64)
		if err != nil {
			return defaultValue
		}

		return time.Minute * time.Duration(int64Value)
	}

	if strings.Contains(value, "h") {
		before, _, _ := strings.Cut(value, "h")
		int64Value, err := strconv.ParseInt(before, 10, 64)
		if err != nil {
			return defaultValue
		}

		return time.Hour * time.Duration(int64Value)
	}

	before, _, found := strings.Cut(value, "d")
	if !found {
		return defaultValue
	}

	int64Value, err := strconv.ParseInt(before, 10, 64)
	if err != nil {
		return defaultValue
	}

	return time.Hour * 24 * time.Duration(int64Value)
}

func (Config) getInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

/*
 * func (Config) getInt64(key string, defaultValue int64) int64 {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value == "" {
		return defaultValue
	}

	int64Value, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}

	return int64Value
 }
*/

func (Config) getBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

/*
 * func (Config) getFloat64(key string, defaultValue float64) float64 {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value == "" {
		return defaultValue
	}

	float64Value, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}

	return float64Value
 }
*/

func (Config) getSlice(key string, defaultValue ...string) []string {
	value := strings.TrimSpace(os.Getenv(strings.ToUpper(key)))
	if value == "" {
		return defaultValue
	}

	keySplit := ","

	if !strings.Contains(value, keySplit) {
		return []string{value}
	}

	return strings.Split(value, keySplit)
}
