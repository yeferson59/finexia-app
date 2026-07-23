package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/identity"
	"github.com/yeferson59/finexia-app/internal/platform/httpx"
)

// ListSessions returns the user's live sessions, flagging the one that issued
// the current request so the client can distinguish "this device".
func (s *Service) ListSessions(ctx context.Context, userID uuid.UUID, currentToken string) ([]ActiveSessionDTO, error) {
	sessions, err := s.stores.Sessions.ListSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]ActiveSessionDTO, 0, len(sessions))
	for _, sess := range sessions {
		result = append(result, ActiveSessionDTO{
			ID:           sess.ID,
			IPAddress:    sess.IPAddress,
			UserAgent:    sess.UserAgent,
			Location:     sess.Location,
			CreatedAt:    sess.CreatedAt,
			LastActiveAt: sess.UpdatedAt,
			ExpiresAt:    sess.ExpiresAt,
			Current:      sess.Token == currentToken,
		})
	}
	return result, nil
}

// RevokeSession terminates one of the user's other sessions. The current
// session must be closed through logout so the refresh cookie is cleared too.
func (s *Service) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID, currentToken string) error {
	sessions, err := s.stores.Sessions.ListSessionsByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, sess := range sessions {
		if sess.ID != sessionID {
			continue
		}
		if sess.Token == currentToken {
			return errors.New("invalid session: use logout to close the current session")
		}
		_, err := s.revokeSessions(ctx, userID, []identity.Session{sess})
		return err
	}

	return httpx.AsNotFound(errors.New("not found session"))
}

// RevokeOtherSessions terminates every session except the one making the
// request and returns how many were closed.
func (s *Service) RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentToken string) (int64, error) {
	sessions, err := s.stores.Sessions.ListSessionsByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}

	others := make([]identity.Session, 0, len(sessions))
	for _, sess := range sessions {
		if sess.Token != currentToken {
			others = append(others, sess)
		}
	}
	if len(others) == 0 {
		return 0, nil
	}

	return s.revokeSessions(ctx, userID, others)
}

// revokeSessions deletes the given sessions and purges every cache entry that
// could keep them alive: the validated-access-token entries, the refresh-token
// entries, and the family revocation markers (set BEFORE the delete cascades,
// mirroring Logout).
func (s *Service) revokeSessions(ctx context.Context, userID uuid.UUID, sessions []identity.Session) (int64, error) {
	ids := make([]uuid.UUID, 0, len(sessions))
	for _, sess := range sessions {
		ids = append(ids, sess.ID)
	}

	if hashes, familyIDs, err := s.stores.RefreshTokens.GetRefreshTokensBySessionIDs(ctx, userID, ids); err == nil {
		for _, hash := range hashes {
			_ = s.storage.DeleteWithContext(ctx, refreshCacheKey(hash))
		}
		seen := make(map[uuid.UUID]struct{}, len(familyIDs))
		for _, familyID := range familyIDs {
			if _, ok := seen[familyID]; ok {
				continue
			}
			seen[familyID] = struct{}{}
			_ = s.storage.SetWithContext(ctx, revokedFamilyCacheKey(familyID), []byte("1"), s.cfg.JWTRefreshDuration)
		}
	}

	for _, sess := range sessions {
		_ = s.storage.DeleteWithContext(ctx, validateTokenCacheKey(sess.Token))
	}

	return s.stores.Sessions.DeleteSessionsByIDs(ctx, userID, ids)
}
