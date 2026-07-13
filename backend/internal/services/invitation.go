package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/entities"
	"github.com/yeferson59/finexia-app/internal/marketing"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
	"github.com/yeferson59/finexia-app/internal/platform/mail"
	"github.com/yeferson59/finexia-app/pkg/helpers"
)

// allowedInviteRoles is the whitelist an admin may assign when inviting. It
// stops an arbitrary role string from being injected into the invitation and,
// through it, into a real account.
var allowedInviteRoles = map[string]bool{
	"customer": true,
	"admin":    true,
}

// Exported so handlers can map each failure to a precise HTTP status and
// message instead of pattern-matching error strings.
var (
	ErrInvitationInvalid = errors.New("invalid invitation")
	ErrInvitationExpired = errors.New("invitation expired")
)

// CreateInvitation issues (or re-issues) an invitation for an email and emails
// the recipient a single-use link. The raw token is returned only inside that
// email; only its hash is persisted. Inviting an address that already has an
// account is refused so the flow can never overwrite existing credentials.
func (s *Services) CreateInvitation(ctx context.Context, email, name, role string, invitedBy uuid.UUID) (entities.Invitation, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return entities.Invitation{}, errors.New("invalid email")
	}

	if role == "" {
		role = "customer"
	}
	if !allowedInviteRoles[role] {
		return entities.Invitation{}, errors.New("invalid role")
	}

	if _, err := s.repos.GetUserByEmail(ctx, email); err == nil {
		return entities.Invitation{}, errors.New("user already exists")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultNameFromEmail(email)
	} else {
		name = helpers.NormalizateNames(name)
	}

	raw, hash, err := generateRefreshToken()
	if err != nil {
		return entities.Invitation{}, err
	}

	var inviter *uuid.UUID
	if invitedBy != uuid.Nil {
		inviter = &invitedBy
	}

	expiresAt := time.Now().UTC().Add(s.cfg.InvitationExpiry)

	inv, err := s.repos.CreateInvitation(ctx, email, name, role, hash, inviter, expiresAt)
	if err != nil {
		return entities.Invitation{}, err
	}

	// Advance the waitlist funnel (pending -> invited) if the person was on the
	// list; a miss is fine, so failures never block the invitation.
	if err := s.repos.SetWaitlistInvited(ctx, email); err != nil {
		s.log.Error(ctx, "create invitation: failed to update waitlist", logger.Err(err))
	}

	s.sendInvitationEmail(ctx, inv, raw)

	return inv, nil
}

// ResendInvitation rotates the token of an existing pending invitation and
// emails the new link, invalidating the previous one.
func (s *Services) ResendInvitation(ctx context.Context, id, invitedBy uuid.UUID) (entities.Invitation, error) {
	existing, err := s.repos.GetInvitationByID(ctx, id)
	if err != nil {
		return entities.Invitation{}, err
	}
	if existing.Status() == "accepted" {
		return entities.Invitation{}, errors.New("invitation already accepted")
	}
	if existing.Status() == "revoked" {
		return entities.Invitation{}, errors.New("invitation revoked")
	}

	// Reuse CreateInvitation so the token rotation and email go through the same
	// path; the partial unique index makes this update the existing row in place.
	return s.CreateInvitation(ctx, existing.Email, existing.Name, existing.Role, invitedBy)
}

func (s *Services) ListInvitations(ctx context.Context, offset, limit uint) ([]entities.Invitation, uint, error) {
	return s.repos.ListInvitations(ctx, offset, limit)
}

func (s *Services) RevokeInvitation(ctx context.Context, id uuid.UUID) error {
	return s.repos.RevokeInvitation(ctx, id)
}

func (s *Services) ListWaitlist(ctx context.Context, offset, limit uint) ([]marketing.Waitlist, uint, error) {
	return s.repos.ListWaitlist(ctx, offset, limit)
}

// ValidateInvitation resolves a raw token to its invitation and enforces that it
// is still redeemable, so the accept page can prefill the email and reject dead
// links before asking for a password.
func (s *Services) ValidateInvitation(ctx context.Context, rawToken string) (entities.Invitation, error) {
	inv, err := s.lookupPendingInvitation(ctx, rawToken)
	if err != nil {
		return entities.Invitation{}, err
	}
	return inv, nil
}

// AcceptInvitation consumes a valid invitation and provisions the account with
// the password the invitee chose. The heavy lifting (create user + account,
// mark consumed, advance waitlist) happens in a single repository transaction.
func (s *Services) AcceptInvitation(ctx context.Context, rawToken, name, password string) (entities.User, error) {
	inv, err := s.lookupPendingInvitation(ctx, rawToken)
	if err != nil {
		return entities.User{}, err
	}

	finalName := strings.TrimSpace(name)
	if finalName == "" {
		finalName = inv.Name
	} else {
		finalName = helpers.NormalizateNames(finalName)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return entities.User{}, err
	}

	return s.repos.AcceptInvitation(ctx, inv.ID, finalName, inv.Email, inv.Role, string(hashed))
}

// lookupPendingInvitation hashes the raw token, fetches the invitation, and
// rejects anything not currently redeemable. The hash comparison happens in the
// database via the unique token_hash column, so the raw token is never stored.
func (s *Services) lookupPendingInvitation(ctx context.Context, rawToken string) (entities.Invitation, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return entities.Invitation{}, ErrInvitationInvalid
	}

	hash, err := hashRefreshToken(rawToken)
	if err != nil {
		return entities.Invitation{}, ErrInvitationInvalid
	}

	inv, err := s.repos.GetInvitationByHash(ctx, hash)
	if err != nil {
		return entities.Invitation{}, ErrInvitationInvalid
	}

	switch inv.Status() {
	case "accepted", "revoked":
		return entities.Invitation{}, ErrInvitationInvalid
	case "expired":
		return entities.Invitation{}, ErrInvitationExpired
	}

	return inv, nil
}

// sendInvitationEmail delivers the invitation link. Best-effort and async: a
// mail hiccup must not fail the admin's request, and the invitation can always
// be resent.
func (s *Services) sendInvitationEmail(ctx context.Context, inv entities.Invitation, rawToken string) {
	if s.mail == nil {
		return
	}

	inviteURL := fmt.Sprintf("%s/auth/accept-invite?token=%s",
		strings.TrimRight(s.cfg.FrontendURL, "/"), url.QueryEscape(rawToken))

	data := mail.InvitationData{
		UserName:  inv.Name,
		InviteURL: inviteURL,
		ExpiresIn: humanizeExpiry(s.cfg.InvitationExpiry),
	}

	go func() {
		if err := s.mail.SendInvitation(inv.Email, data); err != nil {
			s.log.Error(ctx, "failed to send invitation email", logger.Err(err))
		}
	}()
}

// defaultNameFromEmail derives a friendly greeting from the local part of an
// email when the admin invites without providing a name.
func defaultNameFromEmail(email string) string {
	local, _, _ := strings.Cut(email, "@")
	local = strings.NewReplacer(".", " ", "_", " ", "-", " ").Replace(local)
	local = strings.TrimSpace(local)
	if local == "" {
		return "Nuevo usuario"
	}
	return helpers.NormalizateNames(local)
}

// humanizeExpiry renders the invitation lifetime for the email copy, preferring
// whole days when the duration divides evenly.
func humanizeExpiry(d time.Duration) string {
	hours := int(d.Hours())
	if hours <= 0 {
		return "poco tiempo"
	}
	if hours%24 == 0 {
		days := hours / 24
		if days == 1 {
			return "1 día"
		}
		return fmt.Sprintf("%d días", days)
	}
	if hours == 1 {
		return "1 hora"
	}
	return fmt.Sprintf("%d horas", hours)
}
