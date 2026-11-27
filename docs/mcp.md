# Model Context Protocol

Cassocial provides a remote Streamable HTTP MCP endpoint. It exposes the
project's OpenAPI operations as tools and dispatches tool calls through the
same in-process API router used by REST clients.

## Endpoint and authentication

Use the configured API version for the canonical endpoint. The two aliases
always target that same version and do not redirect.

| Endpoint | Purpose |
|----------|---------|
| `/api/{api_version}/server/mcp` | Canonical endpoint |
| `/api/mcp` | Unversioned API alias |
| `/mcp` | Root alias |

Every request requires `Authorization: Bearer ...`. MCP uses existing API
tokens and adds no token type or scope. For a Regular User, create a dedicated
token at `/users/tokens`, choose `read` or `read-write`, and set an expiry. A
token can be revoked independently. A `read` token sees read tools; a
`read-write` token cannot use delete or Server Admin tools; `global` grants
the operations the token owner may perform. Server Admin tokens are for
operator tools, not general clients.

MCP ignores browser cookies; a valid Bearer token is required on every
request. Each POST must include `MCP-Protocol-Version: 2026-07-28` and
`Mcp-Method` matching the JSON-RPC method. The protocol-version header must
match `params._meta.io.modelcontextprotocol/protocolVersion`. Calls to
`tools/call` also require `Mcp-Name` to match `params.name`. Missing or
mismatched headers return a JSON-RPC error with HTTP 400. The endpoint rejects
the older `initialize` handshake and protocol versions other than
`2026-07-28`. Unknown methods return HTTP 404; `GET` and `DELETE` return HTTP
405. The caller's token scope filters both the `tools/list` response and
`tools/call` authorization.

## Connect a client

Set the token in your shell without putting it in a command history:

```sh
export MCP_TOKEN='your-existing-api-token'
```

Add the endpoint to Claude Code, replacing the host and API version with the
values shown on `/server/help/mcp`:

```sh
claude mcp add --transport http cassocial https://YOUR_SERVER/api/v1/server/mcp --header "Authorization: Bearer $MCP_TOKEN"
```

A generic client configuration uses the same endpoint and header:

```json
{
  "mcpServers": {
    "cassocial": {
      "type": "http",
      "url": "https://YOUR_SERVER/api/v1/server/mcp",
      "headers": {
        "Authorization": "Bearer $MCP_TOKEN"
      }
    }
  }
}
```

Verify the connection by requesting the available tools:

```sh
curl -sS -X POST 'https://YOUR_SERVER/api/v1/server/mcp' \
  -H "Authorization: Bearer $MCP_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'MCP-Protocol-Version: 2026-07-28' \
  -H 'Mcp-Method: tools/list' \
  --data '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28"}}}'
```

The live connection guide includes the request-specific, resolved URL and a
tool list. `tools/call` requests also send `Mcp-Name` with the tool name.
Browser origins are validated; command-line clients without an `Origin`
header are unaffected. The endpoint does not perform an initialize handshake,
keep sessions, or accept `GET`/`DELETE` requests.

When an `Origin` header is present, its origin must match the request's own
origin, a configured public origin, or an explicit CORS allow-list entry. A
wildcard CORS default does not authorize MCP origins. An unapproved origin
returns HTTP 403. The endpoint reuses normal API token validation, ownership
checks, rate limits, and audit logging; MCP requests are marked with source
`mcp`.

## Cassocial tools

Tools cover public profile reads and authenticated operations on the caller's
own profiles, links, analytics, imports, exports, and short links. Examples
include `getPublicProfile`, `listPublicProfileLinks`, `listUserProfiles`,
`createUserProfile`, `createProfileLink`, `getProfileAnalytics`, and
`createUserShortlink`. The help page's live list is generated from the API
definition. Cassocial does not currently expose MCP resources or prompts.

The caller's token permissions and the normal API ownership checks apply to
every tool call. MCP traffic uses the API rate limits and is identified as
MCP in audit records.

## Project guide

Server Admins can edit the Cassocial-specific section shown on
`/server/help/mcp` at `/server/{admin_path}/pages/help`. The guide is saved to
`server.pages.help.mcp` in `yaml.d/pages.yml` and becomes active immediately.
Keep the content aligned with the real tools and project data described in
`IDEA.md`.

## Troubleshooting

| Status | Meaning | What to check |
|--------|---------|---------------|
| `400` | Unsupported protocol version or missing/mismatched MCP headers | Use protocol version `2026-07-28`; match the method and tool-name headers to the JSON-RPC body. |
| `401` | Missing, invalid, expired, or revoked Bearer token | Create or refresh an API token and send it in the `Authorization` header. |
| `403` | Unapproved browser origin or insufficient token scope | Use an allowed origin and a token whose scope includes the requested tool. |
| `404` | Unknown MCP method | Use a supported JSON-RPC method such as `tools/list` or `tools/call`. |
| `405` | HTTP method is not supported | Send MCP requests with `POST`. |
