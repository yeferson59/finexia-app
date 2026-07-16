package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yeferson59/finexia-app/internal/identity"
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

// CreateInvitation issues (or re-issues) an invitation for an email and emails
// the recipient a single-use link. The raw token is returned only inside that
// email; only its hash is persisted. Inviting an address that already has an
// account is refused so the flow can never overwrite existing credentials.
func (s *Service) CreateInvitation(ctx context.Context, email, name, role string, invitedBy uuid.UUID) (Invitation, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return Invitation{}, errors.New("invalid email")
	}

	if role == "" {
		role = "customer"
	}
	if !allowedInviteRoles[role] {
		return Invitation{}, errors.New("invalid role")
	}

	if _, err := s.stores.Accounts.GetUserByEmail(ctx, email); err == nil {
		return Invitation{}, errors.New("user already exists")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultNameFromEmail(email)
	} else {
		name = helpers.NormalizateNames(name)
	}

	raw, hash, err := generateRefreshToken()
	if err != nil {
		return Invitation{}, err
	}

	var inviter *uuid.UUID
	if invitedBy != uuid.Nil {
		inviter = &invitedBy
	}

	expiresAt := time.Now().UTC().Add(s.cfg.InvitationExpiry)

	inv, err := s.stores.Invitations.CreateInvitation(ctx, email, name, role, hash, inviter, expiresAt)
	if err != nil {
		return Invitation{}, err
	}

	// Advance the waitlist funnel (pending -> invited) if the person was on the
	// list; a miss is fine, so failures never block the invitation.
	if err := s.stores.Waitlist.SetWaitlistInvited(ctx, email); err != nil {
		s.log.Error(ctx, "create invitation: failed to update waitlist", logger.Err(err))
	}

	s.sendInvitationEmail(ctx, inv, raw)

	return inv, nil
}

// ResendInvitation rotates the token of an existing pending invitation and
// emails the new link, invalidating the previous one.
func (s *Service) ResendInvitation(ctx context.Context, id, invitedBy uuid.UUID) (Invitation, error) {
	existing, err := s.stores.Invitations.GetInvitationByID(ctx, id)
	if err != nil {
		return Invitation{}, err
	}
	if existing.Status() == "accepted" {
		return Invitation{}, errors.New("invitation already accepted")
	}
	if existing.Status() == "revoked" {
		return Invitation{}, errors.New("invitation revoked")
	}

	// Reuse CreateInvitation so the token rotation and email go through the same
	// path; the partial unique index makes this update the existing row in place.
	return s.CreateInvitation(ctx, existing.Email, existing.Name, existing.Role, invitedBy)
}

func (s *Service) ListInvitations(ctx context.Context, offset, limit uint) ([]Invitation, uint, error) {
	return s.stores.Invitations.ListInvitations(ctx, offset, limit)
}

func (s *Service) RevokeInvitation(ctx context.Context, id uuid.UUID) error {
	return s.stores.Invitations.RevokeInvitation(ctx, id)
}

func (s *Service) ListWaitlist(ctx context.Context, offset, limit uint) ([]marketing.Waitlist, uint, error) {
	return s.stores.Waitlist.ListWaitlist(ctx, offset, limit)
}

// ValidateInvitation resolves a raw token to its invitation and enforces that it
// is still redeemable, so the accept page can prefill the email and reject dead
// links before asking for a password.
func (s *Service) ValidateInvitation(ctx context.Context, rawToken string) (Invitation, error) {
	inv, err := s.lookupPendingInvitation(ctx, rawToken)
	if err != nil {
		return Invitation{}, err
	}
	return inv, nil
}

// AcceptInvitation consumes a valid invitation and provisions the account with
// the password the invitee chose. The heavy lifting (create user + account,
// mark consumed, advance waitlist) happens in a single repository transaction.
func (s *Service) AcceptInvitation(ctx context.Context, rawToken, name, password string) (identity.User, error) {
	inv, err := s.lookupPendingInvitation(ctx, rawToken)
	if err != nil {
		return identity.User{}, err
	}

	finalName := strings.TrimSpace(name)
	if finalName == "" {
		finalName = inv.Name
	} else {
		finalName = helpers.NormalizateNames(finalName)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return identity.User{}, err
	}

	return s.stores.Invitations.AcceptInvitation(ctx, inv.ID, finalName, inv.Email, inv.Role, string(hashed))
}

// lookupPendingInvitation hashes the raw token, fetches the invitation, and
// rejects anything not currently redeemable. The hash comparison happens in the
// database via the unique token_hash column, so the raw token is never stored.
func (s *Service) lookupPendingInvitation(ctx context.Context, rawToken string) (Invitation, error) {
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return Invitation{}, ErrInvitationInvalid
	}

	hash, err := hashRefreshToken(rawToken)
	if err != nil {
		return Invitation{}, ErrInvitationInvalid
	}

	inv, err := s.stores.Invitations.GetInvitationByHash(ctx, hash)
	if err != nil {
		return Invitation{}, ErrInvitationInvalid
	}

	switch inv.Status() {
	case "accepted", "revoked":
		return Invitation{}, ErrInvitationInvalid
	case "expired":
		return Invitation{}, ErrInvitationExpired
	}

	return inv, nil
}

// sendInvitationEmail delivers the invitation link. Best-effort and async: a
// mail hiccup must not fail the admin's request, and the invitation can always
// be resent.
func (s *Service) sendInvitationEmail(ctx context.Context, inv Invitation, rawToken string) {
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
