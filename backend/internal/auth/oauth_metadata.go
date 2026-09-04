package auth

// The two discovery documents. They are the whole reason a remote MCP client
// can connect at all: a client that gets a 401 from /mcp reads the
// WWW-Authenticate header, fetches the protected-resource document it names,
// follows that to the authorization server's, and only then knows where to
// register and where to send the user. Nothing else in this module is
// reachable without them.

// OAuthProtectedResourceMetadata is RFC 9728: what /mcp is and who can issue
// tokens for it.
type OAuthProtectedResourceMetadata struct {
	Resource               string   `json:"resource"`
	AuthorizationServers   []string `json:"authorization_servers"`
	ScopesSupported        []string `json:"scopes_supported"`
	BearerMethodsSupported []string `json:"bearer_methods_supported"`
	ResourceName           string   `json:"resource_name"`
	ResourceDocumentation  string   `json:"resource_documentation,omitempty"`
}

// OAuthServerMetadata is RFC 8414: where the endpoints are and what this
// server will accept at them.
//
// Everything advertised here is enforced somewhere else in this package, which
// is the only property that makes the document worth publishing: a client
// configures itself from it and never asks again, so a field that overstates
// what is supported becomes a client that fails at the endpoint instead of at
// discovery.
type OAuthServerMetadata struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	RegistrationEndpoint              string   `json:"registration_endpoint"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	// RFC 8707. Advertised because the MCP specification has clients send a
	// resource indicator, and a client that cannot tell whether the server
	// understands it has to guess.
	ResourceIndicatorsSupported bool     `json:"resource_indicators_supported"`
	ResponseModesSupported      []string `json:"response_modes_supported"`
	ServiceDocumentation        string   `json:"service_documentation,omitempty"`
}

// ProtectedResourceMetadata describes /mcp.
func (s *service) ProtectedResourceMetadata() OAuthProtectedResourceMetadata {
	return OAuthProtectedResourceMetadata{
		Resource:             s.mcpResourceURI(),
		AuthorizationServers: []string{s.issuer()},
		ScopesSupported:      []string{MCPScope},
		// "header" only: a token in a query string ends up in access logs, in
		// Referer headers and in browser history, and this one reads a person's
		// entire portfolio.
		BearerMethodsSupported: []string{"header"},
		ResourceName:           "Finexia",
	}
}

// ServerMetadata describes this authorization server.
func (s *service) ServerMetadata() OAuthServerMetadata {
	issuer := s.issuer()

	return OAuthServerMetadata{
		Issuer:                        issuer,
		AuthorizationEndpoint:         issuer + "/oauth/authorize",
		TokenEndpoint:                 issuer + "/oauth/token",
		RegistrationEndpoint:          issuer + "/oauth/register",
		ScopesSupported:               []string{MCPScope},
		ResponseTypesSupported:        []string{"code"},
		GrantTypesSupported:           []string{"authorization_code", "refresh_token"},
		CodeChallengeMethodsSupported: []string{oauthCodeChallengeS256},
		// "none" first because it is what a client that cannot hold a secret
		// should pick, and the order is the only hint the document carries.
		TokenEndpointAuthMethodsSupported: []string{"none", "client_secret_post", "client_secret_basic"},
		ResourceIndicatorsSupported:       true,
		ResponseModesSupported:            []string{"query"},
	}
}

// mcpChallengeHeader is the WWW-Authenticate value /mcp answers a missing or
// rejected credential with.
//
// The resource_metadata parameter is the entry point to everything above: RFC
// 9728 §5.1 has the client read it off this header rather than guessing at
// well-known paths, and without it a client that receives a 401 knows only that
// it was refused. That is precisely the state this server was in before —
// answering correctly, and telling nobody how to proceed.
func (s *service) mcpChallengeHeader() string {
	return `Bearer realm="finexia", resource_metadata="` + s.issuer() + `/.well-known/oauth-protected-resource/mcp"`
}
