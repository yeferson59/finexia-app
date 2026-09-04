package mcp

import (
	"context"
	"net/http"
	"time"

	"uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/yeferson59/finexia-app/internal/platform/httpx"
	"github.com/yeferson59/finexia-app/internal/platform/logger"
)

// toolTimeout bounds one tool call.
//
// It exists because the deadline the rest of the app relies on does not reach
// here: the MCP session is connected with the HTTP request's context for its
// *values*, but its cancellation is deliberately decoupled, so a tool handler
// would otherwise query the database with no deadline at all. Thirty seconds
// matches the scheduler's default per-attempt budget and is far longer than any
// read below needs.
const toolTimeout = 30 * time.Second

// caller is the authenticated identity, as the tools need it: the user whose
// portfolios are being read, and whether they moderate the shared asset
// catalog. It is the only per-request state the MCP layer carries.
type caller struct {
	userID uuid.UUID
	role   string
}

// isAdmin reports whether this caller sees the whole asset catalog rather than
// the curated rows plus their own.
func (c caller) isAdmin() bool {
	return c.role == httpx.RoleAdmin
}

type callerKey struct{}

func withCaller(ctx context.Context, c caller) context.Context {
	return context.WithValue(ctx, callerKey{}, c)
}

func callerFrom(ctx context.Context) (caller, bool) {
	c, ok := ctx.Value(callerKey{}).(caller)

	return c, ok
}

// bindCaller copies the identity the auth middleware wrote into the request
// locals onto the request's context.Context.
//
// This is the bridge across the transport boundary. The MCP handler is a
// net/http handler behind the Fiber adaptor, so it never sees a fiber.Ctx and
// cannot call httpx.Identity for itself; adaptor.HTTPHandlerWithContext
// forwards whatever c.Context() holds, which makes the context the one place
// the caller can travel through.
func bindCaller(c fiber.Ctx) error {
	userID, _, role, err := httpx.Identity(c)
	if err != nil {
		return httpx.Unauthorized(c, "Invalid identity", err.Error())
	}

	c.SetContext(withCaller(c, caller{userID: userID, role: role}))

	return c.Next()
}

// transport builds the MCP endpoint: the SDK's streamable HTTP handler, wrapped
// for Fiber.
//
// Neither option is a workaround for the adaptor. fasthttp's adaptor does honour
// Flush since v1.73 — an event stream reaches the client through it — so both
// are choices about what this server needs to be.
//
// Stateless means no session is held in memory between requests, so a second
// replica behind the proxy answers a client's next call as well as the first
// did, with no sticky routing and nothing to lose on a deploy. It is also where
// the spec is going (SEP-2567). Its cost is that the server can issue no
// request of its own — there is no open channel to answer on — and that GET and
// DELETE become 405.
//
// JSONResponse follows from that: with no server->client requests to interleave,
// a POST's event stream would carry exactly one message, which is a JSON body
// wrapped in SSE framing. Plain JSON is the same answer with less to parse.
//
// What this forgoes is progress notifications during a long tool call: they do
// survive stateless mode, but only over an event stream. Dropping JSONResponse
// is all that would take, and no tool here runs long enough to want it.
func (m *Module) transport() http.Handler {
	handler := mcpsdk.NewStreamableHTTPHandler(m.getServer, new(mcpsdk.StreamableHTTPOptions{
		Stateless:    true,
		JSONResponse: true,
	}))

	return handler
}

// getServer returns the MCP server that will answer this request, bound to the
// caller who made it. The SDK calls it once per request, which is what makes a
// per-user server affordable — and necessary: the tools close over the caller,
// so there is no shared server whose tools could read the wrong user's data.
//
// A nil return makes the SDK answer 400. It means the request reached here
// without an identity, which RequireAuth already rules out, so it is a wiring
// bug and is logged as one.
func (m *Module) getServer(r *http.Request) *mcpsdk.Server {
	ctx, ok := adaptor.LocalContextFromHTTPRequest(r)
	if !ok {
		m.log.Error(r.Context(), "mcp: request context missing; route not mounted through adaptor.HTTPHandlerWithContext")

		return nil
	}

	c, ok := callerFrom(ctx)
	if !ok {
		m.log.Error(ctx, "mcp: no authenticated caller in context; route not guarded by bindCaller")

		return nil
	}

	return m.newServer(c)
}

// newServer assembles the tool surface for one caller. The schema cache is
// shared across calls, so the reflection that turns each In/Out type into a
// JSON schema is paid once for the process rather than once per request.
func (m *Module) newServer(c caller) *mcpsdk.Server {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:        serverName,
		Title:       "Finexia",
		Description: "The authenticated user's investment portfolios, holdings and market data.",
		Version:     serverVersion,
	}, &mcpsdk.ServerOptions{
		Instructions: instructions,
		SchemaCache:  m.schemas,
	})

	m.addPortfolioTools(server, c)
	m.addMarketTools(server, c)

	return server
}

// readTool registers one read-only tool.
//
// Every tool in this package goes through it, which is what keeps three things
// from being decided per tool: the deadline (see toolTimeout), the annotations
// a client uses to know the call is safe to make unprompted, and the shape of
// the handler — take typed input, return typed output, let the SDK infer both
// schemas and marshal the result.
func readTool[In, Out any](s *mcpsdk.Server, name, title, description string, h func(context.Context, In) (Out, error)) {
	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        name,
		Description: description,
		Annotations: &mcpsdk.ToolAnnotations{
			Title:        title,
			ReadOnlyHint: true,
			// Reading the same holdings twice changes nothing, and everything
			// these tools reach lives in this application's own database — the
			// market-data providers are called by the scheduler, never here.
			IdempotentHint: true,
			OpenWorldHint:  new(false),
		},
	}, func(ctx context.Context, _ *mcpsdk.CallToolRequest, in In) (*mcpsdk.CallToolResult, Out, error) {
		ctx, cancel := context.WithTimeout(ctx, toolTimeout)
		defer cancel()

		out, err := h(ctx, in)
		if err != nil {
			var zero Out

			return nil, zero, err
		}

		return nil, out, nil
	})
}

// logToolError records a failed tool call and returns the error unchanged. The
// SDK turns it into a tool error the model can read, which is the right answer
// for the client; this is what keeps it from also being invisible to us.
func (m *Module) logToolError(ctx context.Context, tool string, c caller, err error) error {
	m.log.Error(ctx, "mcp: tool call failed",
		logger.Str("tool", tool),
		logger.Str("user_id", c.userID.String()),
		logger.Err(err),
	)

	return err
}
