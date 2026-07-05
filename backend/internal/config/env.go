package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Env struct {
	Environment             string
	Port                    string
	PathMigration           string
	DatabaseURL             string
	CacheURL                string
	JWTSecret               string
	JWTAccessDuration       time.Duration
	JWTRefreshDuration      time.Duration
	RefreshGracePeriod      time.Duration
	MaxLoginAttempts        int
	LoginLockout            time.Duration
	TrustProxy              bool
	TrustedProxies          []string
	CORSEnabled             bool
	CORSOrigin              []string
	AWSAccessKeyID          string
	AWSDefaultRegion        string
	AWSEndpointURL          string
	AWSS3BucketName         string
	AWSSecretAccessKey      string
	ResendAPIKey            string
	EmailFrom               string
	AlphaVantageAPIKey      string
	FinnhubAPIKey           string
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

	return &Env{
		Environment:             c.getString("ENVIRONMENT", "development"),
		Port:                    c.getString("PORT", "8080"),
		PathMigration:           c.getString("PATH_MIGRATION", "file://internal/migrations"),
		DatabaseURL:             c.getString("DATABASE_URL", ""),
		CacheURL:                c.getString("CACHE_URL", ""),
		JWTSecret:               c.getString("JWT_SECRET", "secret"),
		JWTAccessDuration:       c.getDuration("JWT_ACCESS_DURATION", 15*time.Minute),
		JWTRefreshDuration:      c.getDuration("JWT_REFRESH_DURATION", 30*24*time.Hour),
		RefreshGracePeriod:      c.getDuration("JWT_REFRESH_GRACE_PERIOD", 30*time.Second),
		MaxLoginAttempts:        c.getInt("MAX_LOGIN_ATTEMPTS", 5),
		LoginLockout:            c.getDuration("LOGIN_LOCKOUT_DURATION", 15*time.Minute),
		TrustProxy:              c.getBool("TRUST_PROXY", true),
		TrustedProxies:          c.getSlice("TRUSTED_PROXIES"),
		CORSEnabled:             c.getBool("CORS_ENABLED", true),
		CORSOrigin:              c.getSlice("CORS_ORIGIN", "http://localhost:5173"),
		AWSAccessKeyID:          c.getString("AWS_ACCESS_KEY_ID", ""),
		AWSDefaultRegion:        c.getString("AWS_DEFAULT_REGION", ""),
		AWSEndpointURL:          c.getString("AWS_ENDPOINT_URL", ""),
		AWSS3BucketName:         c.getString("AWS_S3_BUCKET_NAME", ""),
		AWSSecretAccessKey:      c.getString("AWS_SECRET_ACCESS_KEY", ""),
		ResendAPIKey:            c.getString("RESEND_API_KEY", ""),
		EmailFrom:               c.getString("EMAIL_FROM", "Finexia <noreply@finexia.me>"),
		AlphaVantageAPIKey:      c.getString("ALPHA_VANTAGE_API_KEY", ""),
		FinnhubAPIKey:           c.getString("FINNHUB_API_KEY", ""),
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
	}
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
