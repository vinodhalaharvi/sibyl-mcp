// Command sibyl-mcp exposes Sibyl's durable researcher–critic convergence
// loop as an MCP tool. Any MCP client (Antigravity, Claude, Cursor) can call
// the `deliberate` tool to run a ConvergeWorkflow and get the converged answer.
//
// This process is an MCP server AND a Temporal client — it does NOT run the
// Sibyl worker. Start the worker separately (in the sibyl repo:
// `go run ./cmd/worker`) along with a Temporal dev server. sibyl-mcp only
// submits workflows and waits for their results.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.temporal.io/sdk/client"

	"github.com/vinodhalaharvi/sibyl/agent"
)

// DeliberateInput is the tool's argument schema, surfaced to the calling LLM.
type DeliberateInput struct {
	Question  string `json:"question" jsonschema:"the question or task to deliberate on"`
	MaxRounds int    `json:"max_rounds,omitempty" jsonschema:"max researcher/critic rounds (default 3)"`
}

// DeliberateOutput is the structured result returned to the client.
type DeliberateOutput struct {
	Answer    string `json:"answer" jsonschema:"the converged answer text"`
	Converged bool   `json:"converged" jsonschema:"true if the critic approved before max_rounds"`
	Rounds    int    `json:"rounds" jsonschema:"how many researcher/critic rounds ran"`
}

type server struct {
	tc client.Client
}

func (s *server) deliberate(ctx context.Context, _ *mcp.CallToolRequest, in DeliberateInput) (*mcp.CallToolResult, DeliberateOutput, error) {
	if in.Question == "" {
		return nil, DeliberateOutput{}, fmt.Errorf("question must not be empty")
	}
	rounds := in.MaxRounds
	if rounds <= 0 {
		rounds = 3
	}

	opts := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("sibyl-mcp-%d", time.Now().UnixNano()),
		TaskQueue: agent.TaskQueue,
	}
	we, err := s.tc.ExecuteWorkflow(ctx, opts, "ConvergeWorkflow",
		agent.Question{Text: in.Question, MaxRounds: rounds})
	if err != nil {
		return nil, DeliberateOutput{}, fmt.Errorf("start workflow: %w", err)
	}

	var ans agent.Answer
	if err := we.Get(ctx, &ans); err != nil {
		return nil, DeliberateOutput{}, fmt.Errorf("workflow failed: %w", err)
	}

	out := DeliberateOutput{Answer: ans.Text, Converged: ans.Converged, Rounds: ans.Rounds}
	text := fmt.Sprintf("converged=%v rounds=%d\n\n%s", ans.Converged, ans.Rounds, ans.Text)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, out, nil
}

func main() {
	tc, err := client.Dial(client.Options{HostPort: os.Getenv("TEMPORAL_HOSTPORT")})
	if err != nil {
		log.Fatalln("temporal dial:", err)
	}
	defer tc.Close()

	s := &server{tc: tc}

	srv := mcp.NewServer(&mcp.Implementation{Name: "sibyl-mcp", Version: "v0.1.0"}, nil)
	mcp.AddTool(srv, &mcp.Tool{
		Name: "deliberate",
		Description: "Run Sibyl's durable researcher-critic convergence loop on a question " +
			"and return the converged answer. Use for questions that benefit from " +
			"structured deliberation and self-critique.",
	}, s.deliberate)

	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalln("mcp server:", err)
	}
}
