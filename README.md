# sibyl-mcp

An [MCP](https://modelcontextprotocol.io) server that exposes
[Sibyl](https://github.com/vinodhalaharvi/sibyl)'s durable **researcher–critic
convergence loop** as a single tool, `deliberate`. Any MCP client — Google
Antigravity, Claude, Cursor, Claude Code — can call it to run a question
through Sibyl's `ConvergeWorkflow` and get back the converged answer, with the
full deliberation recorded in Temporal's event history.

## How it fits together

`sibyl-mcp` is an MCP **server** and a Temporal **client**. It does *not* run
the Sibyl worker. Three processes cooperate:

```
MCP client (Antigravity/Claude/Cursor)
        │  stdio (MCP)
        ▼
   sibyl-mcp  ──ExecuteWorkflow──►  Temporal dev server  ◄──polls──  sibyl worker
   (this repo)        "ConvergeWorkflow"                              (sibyl repo)
```

The worker is where the LLM actually runs (scripted / Anthropic / Claude Code /
your own backend). `sibyl-mcp` just submits the workflow and waits for the
result.

## Requirements

- **Go 1.25+** (the official MCP Go SDK requires it)
- The [Temporal CLI](https://docs.temporal.io/cli) for the local dev server
- A checkout of the [sibyl](https://github.com/vinodhalaharvi/sibyl) repo to run
  the worker

## Run it

```bash
# 1. Temporal dev server
temporal server start-dev --db-filename temporal.db --ui-port 8080

# 2. Sibyl worker (in the sibyl repo). -llm scripted needs no API keys.
go run ./cmd/worker -llm scripted

# 3. Build sibyl-mcp (in this repo)
go build -o sibyl-mcp .
```

`sibyl-mcp` speaks MCP over stdio, so a client launches the binary itself.

Optional env: `TEMPORAL_HOSTPORT` (defaults to `localhost:7233`).

## Wire into an MCP client

Point any stdio-capable MCP client at the binary. Example client config:

```json
{
  "mcpServers": {
    "sibyl": {
      "command": "/absolute/path/to/sibyl-mcp"
    }
  }
}
```

The client will then see a `deliberate` tool:

| Field        | Type   | Notes                                            |
| ------------ | ------ | ------------------------------------------------ |
| `question`   | string | the question or task to deliberate on            |
| `max_rounds` | int    | optional, max researcher/critic rounds (default 3) |

It returns the converged answer text plus `converged` and `rounds`. Open
`http://localhost:8080` to watch each deliberation execute live.

## License

MIT
