package auth

import (
	"context"
	"sync"
	"time"

	"uuid"
)

// oauthMemStore is an in-memory OAuthStore.
//
// The other tests in this package hook individual methods, which is right when
// a case is about one call. These are not: an authorization flow is five calls
// that hand each other rows, and stubbing them one at a time would test the
// stubs' agreement rather than the flow's. So this keeps real state, including
// the two behaviours the flow's security rests on — a code can be claimed
// exactly once, and a grant is keyed by (user, client, scope).
type oauthMemStore struct {
	mu       sync.Mutex
	clients  map[string]oauthClient
	requests map[uuid.UUID]pendingAuthorization
	codes    map[string]*authorizationCode
	grants   map[uuid.UUID]*memGrant
	// roles answers the users/roles join the real query does.
	roles map[uuid.UUID]string
}

type memGrant struct {
	id                        uuid.UUID
	userID                    uuid.UUID
	clientID, scope, resource string
	accessHash, refreshHash   string
	accessExpiresAt           time.Time
	refreshExpiresAt          *time.Time
	lastUsedAt                *time.Time
	createdAt                 time.Time
}

func newOAuthMemStore() *oauthMemStore {
	return new(oauthMemStore{
		clients:  map[string]oauthClient{},
		requests: map[uuid.UUID]pendingAuthorization{},
		codes:    map[string]*authorizationCode{},
		grants:   map[uuid.UUID]*memGrant{},
		roles:    map[uuid.UUID]string{},
	})
}

// attach wires every OAuthStore hook of a fakeRepository to this store.
func (m *oauthMemStore) attach(f *fakeRepository) *fakeRepository {
	f.createOAuthClient = m.CreateOAuthClient
	f.getOAuthClient = m.GetOAuthClient
	f.createAuthorizationRequest = m.CreateAuthorizationRequest
	f.getAuthorizationRequest = m.GetAuthorizationRequest
	f.deleteAuthorizationRequest = m.DeleteAuthorizationRequest
	f.createAuthorizationCode = m.CreateAuthorizationCode
	f.consumeAuthorizationCode = m.ConsumeAuthorizationCode
	f.upsertOAuthGrant = m.UpsertOAuthGrant
	f.getGrantByAccessToken = m.GetGrantByAccessToken
	f.getGrantByRefreshToken = m.GetGrantByRefreshToken
	f.rotateGrantTokens = m.RotateGrantTokens
	f.touchOAuthGrant = m.TouchOAuthGrant
	f.listOAuthGrants = m.ListOAuthGrants
	f.deleteOAuthGrant = m.DeleteOAuthGrant
	f.deleteOAuthGrantsForClient = m.DeleteOAuthGrantsForClient
	f.deleteExpiredOAuthRows = m.DeleteExpiredOAuthRows

	return f
}

func (m *oauthMemStore) CreateOAuthClient(_ context.Context, c oauthClient) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[c.ClientID] = c

	return nil
}

func (m *oauthMemStore) GetOAuthClient(_ context.Context, clientID string) (oauthClient, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.clients[clientID]
	if !ok {
		return oauthClient{}, ErrOAuthClientNotFound
	}

	return c, nil
}

func (m *oauthMemStore) CreateAuthorizationRequest(_ context.Context, req pendingAuthorization) (uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	req.ID = uuid.New()
	m.requests[req.ID] = req

	return req.ID, nil
}

func (m *oauthMemStore) GetAuthorizationRequest(_ context.Context, id uuid.UUID) (pendingAuthorization, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	req, ok := m.requests[id]
	if !ok || time.Now().UTC().After(req.ExpiresAt) {
		return pendingAuthorization{}, ErrOAuthRequestNotFound
	}

	return req, nil
}

func (m *oauthMemStore) DeleteAuthorizationRequest(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.requests[id]; !ok {
		return ErrOAuthRequestNotFound
	}

	delete(m.requests, id)

	return nil
}

func (m *oauthMemStore) CreateAuthorizationCode(_ context.Context, codeHash string, c authorizationCode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	c.ID = uuid.New()
	m.codes[codeHash] = &c

	return nil
}

// ConsumeAuthorizationCode reproduces the conditional UPDATE: the first caller
// claims the code, every later one is told it exists but was not claimed.
func (m *oauthMemStore) ConsumeAuthorizationCode(_ context.Context, codeHash string) (authorizationCode, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	code, ok := m.codes[codeHash]
	if !ok {
		return authorizationCode{}, false, ErrOAuthCodeNotFound
	}

	if code.ConsumedAt != nil {
		return *code, false, nil
	}

	now := time.Now().UTC()
	code.ConsumedAt = &now

	return *code, true, nil
}

func (m *oauthMemStore) UpsertOAuthGrant(
	_ context.Context,
	userID uuid.UUID,
	clientID, scope, resource, accessHash string,
	accessExpiresAt time.Time,
	refreshHash string,
	refreshExpiresAt *time.Time,
) (uuid.UUID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, g := range m.grants {
		if g.userID == userID && g.clientID == clientID && g.scope == scope {
			g.accessHash, g.accessExpiresAt = accessHash, accessExpiresAt
			g.refreshHash, g.refreshExpiresAt = refreshHash, refreshExpiresAt

			return g.id, nil
		}
	}

	g := &memGrant{
		id: uuid.New(), userID: userID, clientID: clientID, scope: scope, resource: resource,
		accessHash: accessHash, accessExpiresAt: accessExpiresAt,
		refreshHash: refreshHash, refreshExpiresAt: refreshExpiresAt,
		createdAt: time.Now().UTC(),
	}
	m.grants[g.id] = g

	return g.id, nil
}

func (m *oauthMemStore) GetGrantByAccessToken(_ context.Context, tokenHash string) (oauthGrantIdentity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, g := range m.grants {
		if g.accessHash == tokenHash {
			return oauthGrantIdentity{
				ID: g.id, UserID: g.userID, Role: m.roleOf(g.userID),
				Scope: g.scope, ExpiresAt: g.accessExpiresAt, LastUsedAt: g.lastUsedAt,
			}, nil
		}
	}

	return oauthGrantIdentity{}, ErrOAuthGrantNotFound
}

func (m *oauthMemStore) GetGrantByRefreshToken(_ context.Context, tokenHash string) (grantRefresh, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, g := range m.grants {
		if g.refreshHash != "" && g.refreshHash == tokenHash {
			return grantRefresh{
				ID: g.id, UserID: g.userID, ClientID: g.clientID,
				Scope: g.scope, Resource: g.resource, RefreshExpiresAt: g.refreshExpiresAt,
			}, nil
		}
	}

	return grantRefresh{}, ErrOAuthGrantNotFound
}

func (m *oauthMemStore) RotateGrantTokens(
	_ context.Context,
	id uuid.UUID,
	accessHash string,
	accessExpiresAt time.Time,
	refreshHash string,
	refreshExpiresAt *time.Time,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	g, ok := m.grants[id]
	if !ok {
		return ErrOAuthGrantNotFound
	}

	g.accessHash, g.accessExpiresAt = accessHash, accessExpiresAt
	g.refreshHash, g.refreshExpiresAt = refreshHash, refreshExpiresAt

	return nil
}

func (m *oauthMemStore) TouchOAuthGrant(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if g, ok := m.grants[id]; ok {
		now := time.Now().UTC()
		g.lastUsedAt = &now
	}

	return nil
}

func (m *oauthMemStore) ListOAuthGrants(_ context.Context, userID uuid.UUID) ([]OAuthGrant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	grants := make([]OAuthGrant, 0)

	for _, g := range m.grants {
		if g.userID != userID {
			continue
		}

		grants = append(grants, OAuthGrant{
			ID:         g.id,
			ClientName: m.clients[g.clientID].Name,
			Scopes:     scopeList(g.scope),
			LastUsedAt: g.lastUsedAt,
			CreatedAt:  g.createdAt,
		})
	}

	return grants, nil
}

func (m *oauthMemStore) DeleteOAuthGrant(_ context.Context, userID, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if g, ok := m.grants[id]; ok && g.userID == userID {
		delete(m.grants, id)

		return nil
	}

	return ErrOAuthGrantNotFound
}

func (m *oauthMemStore) DeleteOAuthGrantsForClient(_ context.Context, userID uuid.UUID, clientID string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var deleted int64

	for id, g := range m.grants {
		if g.userID == userID && g.clientID == clientID {
			delete(m.grants, id)
			deleted++
		}
	}

	return deleted, nil
}

func (m *oauthMemStore) DeleteExpiredOAuthRows(_ context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var deleted int64
	now := time.Now().UTC()

	for id, req := range m.requests {
		if now.After(req.ExpiresAt) {
			delete(m.requests, id)
			deleted++
		}
	}

	return deleted, nil
}

// roleOf answers the roles join. Callers must hold the lock.
func (m *oauthMemStore) roleOf(userID uuid.UUID) string {
	if role, ok := m.roles[userID]; ok {
		return role
	}

	return "user"
}

// grantCount reports how many grants exist, for the cases about revocation.
func (m *oauthMemStore) grantCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.grants)
}
