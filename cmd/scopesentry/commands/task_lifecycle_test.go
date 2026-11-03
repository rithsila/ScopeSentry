package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/client"
	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/output"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const testTaskID = "507f1f77bcf86cd799439011"

type taskClient struct {
	client.Client
	name  string
	args  any
	calls int
	err   error
}

func (f *taskClient) CallTool(_ context.Context, name string, args any) (*mcp.CallToolResult, error) {
	f.calls++
	f.name, f.args = name, args
	if f.err != nil {
		return nil, f.err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: `{"success":true,"id":"507f1f77bcf86cd799439011","message":"request accepted"}`}}}, nil
}

func TestTaskLifecycleCommands(t *testing.T) {
	for _, tc := range []struct{ command, tool string }{{"stop", "stop_scan_task"}, {"resume", "resume_scan_task"}} {
		t.Run(tc.command, func(t *testing.T) {
			for _, args := range [][]string{{testTaskID}, {"--id", testTaskID}, {"--id=" + testTaskID}} {
				var stdout, stderr bytes.Buffer
				fake := &taskClient{}
				code := RunTaskCommand(context.Background(), fake, append([]string{tc.command}, args...), &stdout, &stderr)
				if code != output.ExitSuccess {
					t.Fatalf("code=%d stderr=%s", code, stderr.String())
				}
				if fake.calls != 1 || fake.name != tc.tool || !reflect.DeepEqual(fake.args, map[string]any{"id": testTaskID}) {
					t.Fatalf("wrong call: %+v", fake)
				}
				var payload struct {
					Success bool   `json:"success"`
					ID      string `json:"id"`
				}
				if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil || !payload.Success || payload.ID != testTaskID || stderr.Len() != 0 {
					t.Fatalf("stdout=%s stderr=%s", &stdout, &stderr)
				}
			}
			for _, args := range [][]string{nil, {""}, {"   "}, {"bad"}, {strings.Repeat("z", 24)}, {testTaskID, testTaskID}, {"--id", ""}, {"--id", testTaskID, testTaskID}, {"--id", testTaskID, "--id", testTaskID}, {testTaskID + "," + testTaskID}} {
				var stdout, stderr bytes.Buffer
				fake := &taskClient{}
				code := RunTaskCommand(context.Background(), fake, append([]string{tc.command}, args...), &stdout, &stderr)
				if code != output.ExitUsageError || fake.calls != 0 || stdout.Len() != 0 || stderr.Len() == 0 {
					t.Fatalf("args=%q code=%d calls=%d stdout=%s stderr=%s", args, code, fake.calls, &stdout, &stderr)
				}
			}
			for _, backendErr := range []error{errors.New("transport unavailable"), &client.ToolExecutionError{ToolName: tc.tool, Message: "queue unavailable"}} {
				var stdout, stderr bytes.Buffer
				fake := &taskClient{err: backendErr}
				code := RunTaskCommand(context.Background(), fake, []string{tc.command, testTaskID}, &stdout, &stderr)
				if code != output.ExitExecutionError || stdout.Len() != 0 || !strings.Contains(stderr.String(), backendErr.Error()) {
					t.Fatalf("code=%d stdout=%s stderr=%s", code, &stdout, &stderr)
				}
			}
		})
	}
}

func TestCreateTaskAllOnlineNodes(t *testing.T) {
	fake := &taskClient{}
	var stdout, stderr bytes.Buffer
	code := RunTaskCommand(context.Background(), fake, []string{"create", "--name", "test", "--template", testTaskID, "--all-node", "--target", "example.com\nexample.org"}, &stdout, &stderr)
	if code != output.ExitSuccess || fake.name != "create_scan_task" {
		t.Fatalf("code=%d tool=%s stderr=%s", code, fake.name, &stderr)
	}
	args := fake.args.(map[string]any)
	if args["allNode"] != true || args["target"] != "example.com\nexample.org" {
		t.Fatalf("args=%v", args)
	}
	if _, ok := args["node"]; ok {
		t.Fatalf("all-node without explicit nodes must omit node, not send null: %v", args)
	}
}
