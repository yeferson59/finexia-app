// Package mcp serves the caller's own investment data to Model Context
// Protocol clients — Claude Desktop, Claude Code, any MCP-capable agent — over
// the same HTTP surface and the same bearer token as the REST API.
//
// It is a transport, not a domain. Every tool answers with a use case that
// already exists in portfolio or market, read through the interfaces declared
// here; nothing in this package touches the database and nothing decides what a
// number means. A tool that needed either would be a tool whose use case
// belongs in the module that owns the data.
//
// The whole surface is read-only on purpose. An MCP client is a model deciding
// on its own which tool to call, and the blast radius of a wrong read is a
// wrong answer, while the blast radius of a wrong write is a corrupted cost
// basis. Adding a write tool is one function in tools_*.go plus a ReadOnlyHint
// of false — the decision, not the plumbing, is what is deliberately deferred.
package mcp

import (
	"net/http"

	"github.com/gofiber/fiber/v3"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// serverName and serverVersion identify this server to the clients that
// connect. The version is the protocol surface's, not the application's: it
// changes when a tool is added, removed or changes shape, which is the only
// thing a client can observe.
const (
	serverName    = "finexia"
	serverVersion = "1.0.0"
)

// instructions is the server-level hint every client receives on initialize.
// It carries the two facts a model cannot infer from the schemas and gets
// wrong without them: the amounts are decimal strings, and a position with no
// provider price is carried at cost rather than at market.
const instructions = `Finexia exposes the authenticated user's own investment portfolios.

Every monetary amount is a decimal string in the currency named by the
neighbouring currency field — never a float, and never comparable across
currencies. Ask for a currency explicitly when the user names one; otherwise
each portfolio answers in its own base currency.

Positions whose price could not be fetched with the user's own market-data key
are valued at what they cost, which makes their gain exactly zero. The
priceSource field says which, and "cost" must never be reported as a return.`

type Deps struct {
	// Portfolios and Assets are the two modules whose use cases this transport
	// re-serves. Satisfied by *portfolio.service and *market.service.
	Portfolios PortfolioReader
	Assets     MarketReader
	// AuthMiddl provides the route guard. Every tool answers with one user's
	// private holdings, so the endpoint is guarded like the /portfolios family
	// rather than inheriting anything global.
	AuthMiddl authMiddleware
	// Limiter is the shared per-user rate limiter, injected so an MCP client
	// spends the same budget a browser does.
	Limiter fiber.Handler
	Log     logger.Logger
}

type authMiddleware interface {
	RequireAuth() fiber.Handler
}

type Module struct {
	portfolios PortfolioReader
	assets     MarketReader
	log        logger.Logger
	authMiddl  authMiddleware
	limiter    fiber.Handler
	// schemas is shared by every per-request server so the tools' JSON schemas
	// are inferred by reflection once for the process rather than once per
	// call. See newServer for why a server is built per request at all.
	schemas *mcpsdk.SchemaCache
	// handler is the MCP transport, adapted to Fiber. Built once in New: it
	// holds no per-request state of its own, since the caller travels in the
	// request context and the session is stateless.
	handler http.Handler
}

// New builds the module. A missing dependency panics here rather than at the
// first tools/call: it is wiring, so only the composition root can get it
// wrong, and failing at boot is what keeps a misconfigured build from reaching
// production quietly.
func New(deps Deps) *Module {
	switch {
	case deps.AuthMiddl == nil:
		panic("mcp.New: Deps.AuthMiddl is required — /mcp answers with one user's private holdings")
	case deps.Portfolios == nil:
		panic("mcp.New: Deps.Portfolios is required")
	case deps.Assets == nil:
		panic("mcp.New: Deps.Assets is required")
	case deps.Log == nil:
		panic("mcp.New: Deps.Log is required")
	}

	m := new(Module{
		portfolios: deps.Portfolios,
		assets:     deps.Assets,
		log:        deps.Log,
		authMiddl:  deps.AuthMiddl,
		limiter:    httpx.OrPassThrough(deps.Limiter),
		schemas:    mcpsdk.NewSchemaCache(),
	})

	m.handler = m.transport()

	return m
}

// Routes mounts the single MCP endpoint. GET and DELETE are registered
// alongside POST so the SDK answers them with the 405 the spec asks for in
// stateless mode, instead of Fiber answering with a 404 that a client reads as
// "wrong URL".
func (m *Module) Routes(router fiber.Router) {
	endpoint := router.Group("/mcp")

	endpoint.Use(m.authMiddl.RequireAuth(), m.limiter, bindCaller)
	endpoint.Add([]string{fiber.MethodPost, fiber.MethodGet, fiber.MethodDelete}, "", m.handler)
}
