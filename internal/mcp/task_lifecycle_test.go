package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/task/task"
	"github.com/gin-gonic/gin"
	mcpSDK "github.com/modelcontextprotocol/go-sdk/mcp"
)

const lifecycleTestID = "507f1f77bcf86cd799439011"

type lifecycleTaskService struct {
	task.Service
	stopped, started []string
	err              error
	detail           *models.Task
	detailErr        error
}

func (f *lifecycleTaskService) GetTaskDetail(_ *gin.Context, _ string) (*models.Task, error) {
	return f.detail, f.detailErr
}

func (f *lifecycleTaskService) StopTasks(_ *gin.Context, ids []string) error {
	f.stopped = append(f.stopped, ids...)
	return f.err
}

func (f *lifecycleTaskService) StartTasks(_ *gin.Context, ids []string) error {
	f.started = append(f.started, ids...)
	return f.err
}

func useTestDeps(t *testing.T, d *deps) {
	t.Helper()
	old := depsInstance
	// Complete the lazy initializer with the fake, without constructing real services.
	depsOnce.Do(func() { depsInstance = d })
	depsInstance = d
	t.Cleanup(func() {
		depsInstance = old
		if old == nil {
			depsOnce = sync.Once{}
		}
	})
}

func testSession(t *testing.T) (context.Context, *mcpSDK.ClientSession) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	serverTransport, clientTransport := mcpSDK.NewInMemoryTransports()
	serverSession, err := newServer().Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })
	cli := mcpSDK.NewClient(&mcpSDK.Implementation{Name: "test", Version: "1"}, nil)
	session, err := cli.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return ctx, session
}

func resultText(result *mcpSDK.CallToolResult) string {
	var text strings.Builder
	for _, content := range result.Content {
		if c, ok := content.(*mcpSDK.TextContent); ok {
			text.WriteString(c.Text)
		}
	}
	return text.String()
}

func TestTaskLifecycle(t *testing.T) {
	for _, name := range []string{"stop_scan_task", "resume_scan_task"} {
		t.Run(name, func(t *testing.T) {
			for _, tc := range []struct {
				name       string
				args       map[string]any
				serviceErr error
				valid      bool
			}{
				{"accepted", map[string]any{"id": lifecycleTestID}, nil, true},
				{"service failure", map[string]any{"id": lifecycleTestID}, errors.New("queue unavailable"), true},
				{"missing", map[string]any{}, nil, false},
				{"blank", map[string]any{"id": "   "}, nil, false},
				{"empty", map[string]any{"id": ""}, nil, false},
				{"invalid", map[string]any{"id": "not-an-id"}, nil, false},
				{"multiple", map[string]any{"id": lifecycleTestID + "," + lifecycleTestID}, nil, false},
				{"array", map[string]any{"id": []string{lifecycleTestID}}, nil, false},
			} {
				t.Run(tc.name, func(t *testing.T) {
					fake := &lifecycleTaskService{err: tc.serviceErr, detail: &models.Task{Status: 2}}
					useTestDeps(t, &deps{taskService: fake})
					ctx, session := testSession(t)
					result, err := session.CallTool(ctx, &mcpSDK.CallToolParams{Name: name, Arguments: tc.args})
					if err != nil {
						// Required/type schema failures are protocol errors, not handler results.
						if tc.name != "missing" && tc.name != "array" {
							t.Fatalf("tool call failed: %v", err)
						}
						if !strings.Contains(err.Error(), "invalid params") || len(fake.stopped)+len(fake.started) != 0 {
							t.Fatalf("unexpected schema failure: %v", err)
						}
						return
					}
					wantError := !tc.valid || tc.serviceErr != nil
					if result.IsError != wantError {
						t.Fatalf("IsError=%v, text=%s", result.IsError, resultText(result))
					}
					var wantStopped, wantStarted []string
					if tc.valid {
						if name == "stop_scan_task" {
							wantStopped = []string{lifecycleTestID}
						} else {
							wantStarted = []string{lifecycleTestID}
						}
					}
					if !reflect.DeepEqual(fake.stopped, wantStopped) || !reflect.DeepEqual(fake.started, wantStarted) {
						t.Fatalf("wrong forwarding: stopped=%v started=%v", fake.stopped, fake.started)
					}
					if tc.serviceErr != nil && !strings.Contains(resultText(result), "queue unavailable") {
						t.Fatal("service error lost")
					}
					if !wantError {
						var payload struct {
							Success bool   `json:"success"`
							ID      string `json:"id"`
							Message string `json:"message"`
						}
						if err := json.Unmarshal([]byte(resultText(result)), &payload); err != nil {
							t.Fatal(err)
						}
						if !payload.Success || payload.ID != lifecycleTestID || !strings.Contains(payload.Message, "accepted") {
							t.Fatalf("unexpected acceptance: %+v", payload)
						}
					}
				})
			}
		})
	}
}

func TestTaskLifecycleRejectsUnavailableTask(t *testing.T) {
	for _, name := range []string{"stop_scan_task", "resume_scan_task"} {
		for _, tc := range []struct {
			name   string
			detail *models.Task
			err    error
			want   string
		}{
			{"missing", nil, nil, "not found"},
			{"lookup error", nil, errors.New("database unavailable"), "database unavailable"},
			{"completed status", &models.Task{Status: 3}, nil, "completed"},
			{"completed progress", &models.Task{Status: 2, Progress: 100}, nil, "completed"},
			{"running resume", &models.Task{Status: 1}, nil, "paused"},
		} {
			if name == "stop_scan_task" && tc.name == "running resume" {
				continue
			}
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				fake := &lifecycleTaskService{detail: tc.detail, detailErr: tc.err}
				useTestDeps(t, &deps{taskService: fake})
				ctx, session := testSession(t)
				result, err := session.CallTool(ctx, &mcpSDK.CallToolParams{Name: name, Arguments: map[string]any{"id": lifecycleTestID}})
				if err != nil {
					t.Fatal(err)
				}
				if !result.IsError || !strings.Contains(resultText(result), tc.want) || len(fake.stopped)+len(fake.started) != 0 {
					t.Fatalf("unexpected result: %s; stopped=%v started=%v", resultText(result), fake.stopped, fake.started)
				}
			})
		}
	}
}
