package services

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/yeferson59/finexia-app/internal/dtos/auth"
	"github.com/yeferson59/finexia-app/internal/entities"
)

// ListSessions returns the user's live sessions, flagging the one that issued
// the current request so the client can distinguish "this device".
func (s *Services) ListSessions(ctx context.Context, userID uuid.UUID, currentToken string) ([]auth.ActiveSessionDTO, error) {
	sessions, err := s.repos.ListSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]auth.ActiveSessionDTO, 0, len(sessions))
	for _, sess := range sessions {
		result = append(result, auth.ActiveSessionDTO{
			ID:           sess.ID,
			IPAddress:    sess.IPAddress,
			UserAgent:    sess.UserAgent,
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
func (s *Services) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID, currentToken string) error {
	sessions, err := s.repos.ListSessionsByUserID(ctx, userID)
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
		_, err := s.revokeSessions(ctx, userID, []entities.Session{sess})
		return err
	}

	return errors.New("not found session")
}

// RevokeOtherSessions terminates every session except the one making the
// request and returns how many were closed.
func (s *Services) RevokeOtherSessions(ctx context.Context, userID uuid.UUID, currentToken string) (int64, error) {
	sessions, err := s.repos.ListSessionsByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}

	others := make([]entities.Session, 0, len(sessions))
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
func (s *Services) revokeSessions(ctx context.Context, userID uuid.UUID, sessions []entities.Session) (int64, error) {
	ids := make([]uuid.UUID, 0, len(sessions))
	for _, sess := range sessions {
		ids = append(ids, sess.ID)
	}

	if hashes, familyIDs, err := s.repos.GetRefreshTokensBySessionIDs(ctx, userID, ids); err == nil {
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

	return s.repos.DeleteSessionsByIDs(ctx, userID, ids)
}
