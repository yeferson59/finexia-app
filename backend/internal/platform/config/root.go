package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yeferson59/finexia-app/internal/platform/env"
)

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

type EnvConfig struct {
	Environment        string        `env:"ENVIRONMENT" default:"development"`
	Port               string        `env:"PORT" default:"8080"`
	PathMigration      string        `env:"PATH_MIGRATION" default:"file://internal/migrations"`
	DatabaseURL        string        `env:"DATABASE_URL,required"`
	RedisHost          string        `env:"REDIS_HOST" default:"localhost"`
	RedisPort          string        `env:"REDIS_PORT" default:"6379"`
	RedisPassword      string        `env:"REDIS_PASSWORD"`
	RedisDB            int           `env:"REDIS_DB" default:"0"`
	JWTSecret          string        `env:"JWT_SECRET,required"`
	JWTAccessDuration  time.Duration `env:"JWT_ACCESS_DURATION" default:"15m"`
	JWTRefreshDuration time.Duration `env:"JWT_REFRESH_DURATION" default:"30d"`
	RefreshGracePeriod time.Duration `env:"JWT_REFRESH_GRACE_PERIOD" default:"30s"`
	MaxLoginAttempts   int           `env:"MAX_LOGIN_ATTEMPTS" default:"5"`
	LoginLockout       time.Duration `env:"LOGIN_LOCKOUT_DURATION" default:"15m"`
	TrustProxy         bool          `env:"TRUST_PROXY" default:"true"`
	TrustedProxies     []string      `env:"TRUSTED_PROXIES"`
	CORSEnabled        bool          `env:"CORS_ENABLED" default:"true"`
	CORSOrigin         []string      `env:"CORS_ORIGIN" default:"http://localhost:5173"`
	AWSAccessKeyID     string        `env:"AWS_ACCESS_KEY_ID"`
	AWSDefaultRegion   string        `env:"AWS_DEFAULT_REGION"`
	AWSEndpointURL     string        `env:"AWS_ENDPOINT_URL"`
	AWSS3BucketName    string        `env:"AWS_S3_BUCKET_NAME"`
	AWSSecretAccessKey string        `env:"AWS_SECRET_ACCESS_KEY"`
	ResendAPIKey       string        `env:"RESEND_API_KEY"`
	EmailFrom          string        `env:"EMAIL_FROM" default:"Finexia <noreply@finexia.me>"`
	// Market data is BYO-key: each user supplies their own provider key, so
	// the application holds no provider credentials of its own. These two are
	// the keys that seal the users' keys, not keys to any provider.
	MarketKEKKeys           []string      `env:"MARKET_KEK_KEYS,required"`
	MarketKEKActive         string        `env:"MARKET_KEK_ACTIVE" default:"1"`
	PublicURL               string        `env:"PUBLIC_URL" default:"http://localhost:8080"`
	FrontendURL             string        `env:"FRONTEND_URL" default:"http://localhost:5173"`
	InvitationExpiry        time.Duration `env:"INVITATION_EXPIRY" default:"4d"`
	PasswordResetExpiry     time.Duration `env:"PASSWORD_RESET_EXPIRY" default:"1h"`
	EmailVerificationExpiry time.Duration `env:"EMAIL_VERIFICATION_EXPIRY" default:"1d"`
	SelfRegistrationEnabled bool          `env:"SELF_REGISTRATION_ENABLED" default:"false"`
	TwoFactorPendingExpiry  time.Duration `env:"TWO_FACTOR_PENDING_EXPIRY" default:"5m"`
}

func New() (*EnvConfig, error) {
	envConfig := new(EnvConfig{})

	if err := env.LoadParse(envConfig); err != nil {
		return nil, err
	}

	return envConfig, nil
}

// Validate reports the first configuration error that must stop the process.
//
// It covers the settings whose absence is not a degraded mode but a security
// hole: a deployment missing them would run, serve traffic, and be trivially
// compromised. Everything else in Env has a defensible default and is left to
// the module that consumes it.
func (e *EnvConfig) Validate() error {
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
