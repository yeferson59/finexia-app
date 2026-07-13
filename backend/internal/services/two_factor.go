package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/dtos/auth"
	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/internal/repositories"
	"github.com/yeferson59/finexia-app/pkg/helpers"
)

// Exported so the login handler can distinguish "password OK, now ask for the
// TOTP code" from a plain failure, and the 2FA handlers can map each case to
// a precise HTTP response.
var (
	ErrTwoFactorRequired       = errors.New("two-factor authentication required")
	ErrTwoFactorInvalidCode    = errors.New("invalid two-factor code")
	ErrTwoFactorPendingInvalid = errors.New("invalid or expired two-factor session")
	ErrTwoFactorAlreadyEnabled = errors.New("two-factor already enabled")
	ErrTwoFactorNotEnabled     = errors.New("two-factor not enabled")
	ErrTwoFactorSetupMissing   = errors.New("two-factor setup not started")
)

const (
	// twoFactorIssuer labels the entry inside authenticator apps.
	twoFactorIssuer = "Finexia"
	// twoFactorRecoveryCodeCount is how many single-use fallback codes the
	// user receives when enabling 2FA (and when regenerating).
	twoFactorRecoveryCodeCount = 8
	// twoFactorMaxAttempts bounds code guesses per pending login before the
	// user must re-enter their password.
	twoFactorMaxAttempts = 5
)

func twoFactorPendingCacheKey(hash string) string  { return "2fa_pending:" + hash }
func twoFactorAttemptsCacheKey(hash string) string { return "2fa_attempts:" + hash }

// twoFactorUsedStepCacheKey marks a TOTP time step as consumed for a user, so
// an intercepted code cannot be replayed inside its validity window.
func twoFactorUsedStepCacheKey(userID uuid.UUID, step int64) string {
	return fmt.Sprintf("2fa_used:%s:%d", userID, step)
}

// hashTwoFactorRecoveryCode normalizes user input (case, dashes, spaces) and
// hashes it; only hashes are stored or compared.
func hashTwoFactorRecoveryCode(code string) string {
	normalized := strings.ToUpper(strings.NewReplacer("-", "", " ", "").Replace(strings.TrimSpace(code)))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:])
}

// generateTwoFactorRecoveryCodes returns the raw codes (shown to the user
// exactly once) alongside the hashes that get persisted.
func generateTwoFactorRecoveryCodes() (raws, hashes []string, err error) {
	raws = make([]string, 0, twoFactorRecoveryCodeCount)
	hashes = make([]string, 0, twoFactorRecoveryCodeCount)
	for range twoFactorRecoveryCodeCount {
		buf := make([]byte, 5)
		if _, err = rand.Read(buf); err != nil {
			return nil, nil, err
		}
		hexCode := strings.ToUpper(hex.EncodeToString(buf))
		raw := hexCode[:5] + "-" + hexCode[5:]
		raws = append(raws, raw)
		hashes = append(hashes, hashTwoFactorRecoveryCode(raw))
	}
	return raws, hashes, nil
}

// verifyCurrentPassword gates every 2FA management action behind the account
// password, so a hijacked session alone cannot reconfigure 2FA.
func (s *Services) verifyCurrentPassword(ctx context.Context, userID uuid.UUID, password string) error {
	account, err := s.repos.GetAccountByUserID(ctx, userID)
	if err != nil {
		return errors.New("invalid current password")
	}
	if err := account.ComparePassword(password); err != nil {
		return errors.New("invalid current password")
	}
	return nil
}

// getTwoFactor returns the row plus a found flag, folding the not-found case
// so callers don't repeat the sentinel check.
func (s *Services) getTwoFactor(ctx context.Context, userID uuid.UUID) (entities.TwoFactor, bool, error) {
	tf, err := s.repos.GetTwoFactor(ctx, userID)
	if errors.Is(err, repositories.ErrTwoFactorNotFound) {
		return entities.TwoFactor{}, false, nil
	}
	if err != nil {
		return entities.TwoFactor{}, false, err
	}
	return tf, true, nil
}

// TwoFactorStatus reports whether 2FA is enabled (default: disabled), whether
// a setup is mid-flight, and how many recovery codes remain unused.
func (s *Services) TwoFactorStatus(ctx context.Context, userID uuid.UUID) (auth.TwoFactorStatusResponseDTO, error) {
	tf, found, err := s.getTwoFactor(ctx, userID)
	if err != nil {
		return auth.TwoFactorStatusResponseDTO{}, err
	}

	status := auth.TwoFactorStatusResponseDTO{
		Enabled:      found && tf.Enabled,
		PendingSetup: found && !tf.Enabled,
	}

	if status.Enabled {
		count, err := s.repos.CountUnusedTwoFactorRecoveryCodes(ctx, userID)
		if err != nil {
			return auth.TwoFactorStatusResponseDTO{}, err
		}
		status.RecoveryCodesLeft = count
	}

	return status, nil
}

// BeginTwoFactorSetup issues a fresh TOTP secret for the user to load into an
// authenticator app. Nothing is enforced yet: login only starts requiring
// codes after ConfirmTwoFactorSetup proves the app produces valid ones.
func (s *Services) BeginTwoFactorSetup(ctx context.Context, userID uuid.UUID, password string) (auth.TwoFactorSetupResponseDTO, error) {
	if err := s.verifyCurrentPassword(ctx, userID, password); err != nil {
		return auth.TwoFactorSetupResponseDTO{}, err
	}

	tf, found, err := s.getTwoFactor(ctx, userID)
	if err != nil {
		return auth.TwoFactorSetupResponseDTO{}, err
	}
	if found && tf.Enabled {
		return auth.TwoFactorSetupResponseDTO{}, ErrTwoFactorAlreadyEnabled
	}

	user, err := s.repos.GetUserByID(ctx, userID)
	if err != nil {
		return auth.TwoFactorSetupResponseDTO{}, err
	}

	secret, err := helpers.GenerateTOTPSecret()
	if err != nil {
		return auth.TwoFactorSetupResponseDTO{}, err
	}

	if err := s.repos.UpsertTwoFactorSecret(ctx, userID, secret); err != nil {
		return auth.TwoFactorSetupResponseDTO{}, err
	}

	return auth.TwoFactorSetupResponseDTO{
		Secret:     secret,
		OtpauthURL: helpers.BuildOTPAuthURL(twoFactorIssuer, user.Email, secret),
	}, nil
}

// ConfirmTwoFactorSetup verifies the first code from the authenticator app,
// switches 2FA on, and hands back the single-use recovery codes. The raw
// codes exist only in this response; the database keeps hashes.
func (s *Services) ConfirmTwoFactorSetup(ctx context.Context, userID uuid.UUID, code, ipAddress, userAgent string) (auth.TwoFactorEnableResponseDTO, error) {
	tf, found, err := s.getTwoFactor(ctx, userID)
	if err != nil {
		return auth.TwoFactorEnableResponseDTO{}, err
	}
	if !found {
		return auth.TwoFactorEnableResponseDTO{}, ErrTwoFactorSetupMissing
	}
	if tf.Enabled {
		return auth.TwoFactorEnableResponseDTO{}, ErrTwoFactorAlreadyEnabled
	}

	if !s.verifyTOTPWithReplayGuard(ctx, userID, tf.Secret, code) {
		return auth.TwoFactorEnableResponseDTO{}, ErrTwoFactorInvalidCode
	}

	if err := s.repos.EnableTwoFactor(ctx, userID); err != nil {
		return auth.TwoFactorEnableResponseDTO{}, err
	}

	raws, hashes, err := generateTwoFactorRecoveryCodes()
	if err != nil {
		return auth.TwoFactorEnableResponseDTO{}, err
	}
	if err := s.repos.ReplaceTwoFactorRecoveryCodes(ctx, userID, hashes); err != nil {
		return auth.TwoFactorEnableResponseDTO{}, err
	}

	go s.sendTwoFactorAlert(userID, "autenticación en dos pasos activada",
		"La verificación en dos pasos (2FA) fue activada en tu cuenta. A partir de ahora se pedirá un código del autenticador al iniciar sesión.",
		ipAddress, userAgent)

	return auth.TwoFactorEnableResponseDTO{RecoveryCodes: raws}, nil
}

// DisableTwoFactor turns 2FA back off. It demands the password AND a valid
// code (TOTP or recovery), so neither a stolen session nor a stolen password
// alone is enough to strip the protection.
func (s *Services) DisableTwoFactor(ctx context.Context, userID uuid.UUID, password, code, ipAddress, userAgent string) error {
	if err := s.verifyCurrentPassword(ctx, userID, password); err != nil {
		return err
	}

	tf, found, err := s.getTwoFactor(ctx, userID)
	if err != nil {
		return err
	}
	if !found {
		return ErrTwoFactorNotEnabled
	}

	// A pending (never confirmed) setup can be cancelled with just the
	// password: it is not protecting the account yet.
	if tf.Enabled && !s.verifyTwoFactorCode(ctx, userID, tf.Secret, code) {
		return ErrTwoFactorInvalidCode
	}

	if err := s.repos.DeleteTwoFactor(ctx, userID); err != nil {
		return err
	}

	if tf.Enabled {
		go s.sendTwoFactorAlert(userID, "autenticación en dos pasos desactivada",
			"La verificación en dos pasos (2FA) fue desactivada en tu cuenta. Si no fuiste tú, cambia tu contraseña de inmediato.",
			ipAddress, userAgent)
	}

	return nil
}

// RegenerateTwoFactorRecoveryCodes invalidates every old recovery code and
// issues a fresh batch, guarded by password + current code.
func (s *Services) RegenerateTwoFactorRecoveryCodes(ctx context.Context, userID uuid.UUID, password, code string) (auth.TwoFactorEnableResponseDTO, error) {
	if err := s.verifyCurrentPassword(ctx, userID, password); err != nil {
		return auth.TwoFactorEnableResponseDTO{}, err
	}

	tf, found, err := s.getTwoFactor(ctx, userID)
	if err != nil {
		return auth.TwoFactorEnableResponseDTO{}, err
	}
	if !found || !tf.Enabled {
		return auth.TwoFactorEnableResponseDTO{}, ErrTwoFactorNotEnabled
	}

	if !s.verifyTwoFactorCode(ctx, userID, tf.Secret, code) {
		return auth.TwoFactorEnableResponseDTO{}, ErrTwoFactorInvalidCode
	}

	raws, hashes, err := generateTwoFactorRecoveryCodes()
	if err != nil {
		return auth.TwoFactorEnableResponseDTO{}, err
	}
	if err := s.repos.ReplaceTwoFactorRecoveryCodes(ctx, userID, hashes); err != nil {
		return auth.TwoFactorEnableResponseDTO{}, err
	}

	return auth.TwoFactorEnableResponseDTO{RecoveryCodes: raws}, nil
}

// createTwoFactorPending stores a short-lived opaque token that represents
// "password already verified, waiting for the code". Only its hash is cached,
// mirroring how refresh tokens are handled.
func (s *Services) createTwoFactorPending(ctx context.Context, userID uuid.UUID) (string, error) {
	raw, hash, err := generateRefreshToken()
	if err != nil {
		return "", err
	}
	if err := s.storage.SetWithContext(ctx, twoFactorPendingCacheKey(hash), []byte(userID.String()), s.cfg.TwoFactorPendingExpiry); err != nil {
		return "", err
	}
	return raw, nil
}

// CompleteTwoFactorLogin exchanges a pending-login token plus a valid TOTP or
// recovery code for a real session. Attempts are counted per pending token;
// too many failures kill the token and force a fresh password login.
func (s *Services) CompleteTwoFactorLogin(ctx context.Context, rawToken, code, ipAddress, userAgent string) (auth.LoginInternalDTO, error) {
	ipAddress = sanitizeIP(truncate(ipAddress, 45))
	userAgent = truncate(userAgent, 255)

	hash, err := hashRefreshToken(rawToken)
	if err != nil {
		return auth.LoginInternalDTO{}, ErrTwoFactorPendingInvalid
	}

	pendingKey := twoFactorPendingCacheKey(hash)
	data, _ := s.storage.GetWithContext(ctx, pendingKey)
	if len(data) == 0 {
		return auth.LoginInternalDTO{}, ErrTwoFactorPendingInvalid
	}

	userID, err := uuid.Parse(string(data))
	if err != nil {
		_ = s.storage.DeleteWithContext(ctx, pendingKey)
		return auth.LoginInternalDTO{}, ErrTwoFactorPendingInvalid
	}

	attemptsKey := twoFactorAttemptsCacheKey(hash)
	attempts := 0
	if raw, _ := s.storage.GetWithContext(ctx, attemptsKey); len(raw) > 0 {
		attempts, _ = strconv.Atoi(string(raw))
	}
	if attempts >= twoFactorMaxAttempts {
		_ = s.storage.DeleteWithContext(ctx, pendingKey)
		return auth.LoginInternalDTO{}, ErrTwoFactorPendingInvalid
	}

	tf, found, err := s.getTwoFactor(ctx, userID)
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}
	if !found || !tf.Enabled {
		_ = s.storage.DeleteWithContext(ctx, pendingKey)
		return auth.LoginInternalDTO{}, ErrTwoFactorPendingInvalid
	}

	if !s.verifyTwoFactorCode(ctx, userID, tf.Secret, code) {
		attempts++
		_ = s.storage.SetWithContext(ctx, attemptsKey, []byte(strconv.Itoa(attempts)), s.cfg.TwoFactorPendingExpiry)
		if attempts >= twoFactorMaxAttempts {
			_ = s.storage.DeleteWithContext(ctx, pendingKey)
			return auth.LoginInternalDTO{}, ErrTwoFactorPendingInvalid
		}
		return auth.LoginInternalDTO{}, ErrTwoFactorInvalidCode
	}

	_ = s.storage.DeleteWithContext(ctx, pendingKey)
	_ = s.storage.DeleteWithContext(ctx, attemptsKey)

	user, err := s.repos.GetUserByID(ctx, userID)
	if err != nil {
		return auth.LoginInternalDTO{}, err
	}

	return s.issueSession(ctx, user.ID, user.Role.Name, user.Name, user.Email, ipAddress, userAgent)
}

// verifyTwoFactorCode accepts either a 6-digit TOTP code or a recovery code.
func (s *Services) verifyTwoFactorCode(ctx context.Context, userID uuid.UUID, secret, code string) bool {
	compact := strings.NewReplacer("-", "", " ", "").Replace(strings.TrimSpace(code))
	if len(compact) == 6 {
		if _, err := strconv.Atoi(compact); err == nil {
			return s.verifyTOTPWithReplayGuard(ctx, userID, secret, compact)
		}
	}

	return s.repos.ConsumeTwoFactorRecoveryCode(ctx, userID, hashTwoFactorRecoveryCode(code)) == nil
}

// verifyTOTPWithReplayGuard checks the code and refuses to accept the same
// time step twice, closing the window where a shoulder-surfed or intercepted
// code would still be valid.
func (s *Services) verifyTOTPWithReplayGuard(ctx context.Context, userID uuid.UUID, secret, code string) bool {
	ok, step := helpers.VerifyTOTP(secret, code, time.Now().UTC())
	if !ok {
		return false
	}

	usedKey := twoFactorUsedStepCacheKey(userID, step)
	if used, _ := s.storage.GetWithContext(ctx, usedKey); len(used) > 0 {
		return false
	}
	// TTL covers the step's own 30s plus the ±1 step drift window.
	_ = s.storage.SetWithContext(ctx, usedKey, []byte("1"), 2*time.Minute)
	return true
}

// sendTwoFactorAlert notifies the user of 2FA state changes. Like other
// security notices it ignores marketing preferences and is best-effort.
func (s *Services) sendTwoFactorAlert(userID uuid.UUID, event, detail, ipAddress, userAgent string) {
	if s.mail == nil {
		return
	}

	ctx := context.Background()
	user, err := s.repos.GetUserByID(ctx, userID)
	if err != nil {
		return
	}

	ipAddress = sanitizeIP(ipAddress)
	location := s.locateIP(ipAddress)
	if location == "" {
		location = "desconocida"
	}
	if ipAddress == "" {
		ipAddress = "desconocida"
	}
	if userAgent == "" {
		userAgent = "desconocido"
	}

	_ = s.mail.SendSecurityAlert(user.Email, mail.SecurityAlertData{
		UserName:    user.Name,
		Event:       event,
		Detail:      detail,
		IPAddress:   truncate(ipAddress, 45),
		UserAgent:   truncate(userAgent, 255),
		Location:    location,
		When:        time.Now().UTC().Format("02 Jan 2006 15:04 UTC"),
		SecurityURL: s.cfg.FrontendURL + "/dashboard/settings",
	})
}
