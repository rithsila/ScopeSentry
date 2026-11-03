package mcp

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/services/node"
	taskCommon "github.com/Autumn-27/ScopeSentry/internal/services/task/common"
	"github.com/Autumn-27/ScopeSentry/internal/services/task/task"
	"github.com/Autumn-27/ScopeSentry/internal/services/task/template"
	"github.com/gin-gonic/gin"
	mcpSDK "github.com/modelcontextprotocol/go-sdk/mcp"
)

type preflightTaskService struct {
	task.Service
	exists bool
	err    error
}

func (f *preflightTaskService) CheckTaskNameExists(_ *gin.Context, _ string) (bool, error) {
	return f.exists, f.err
}

type preflightTemplateService struct {
	template.Service
	detail *models.ScanTemplate
	err    error
	ids    []string
}

func (f *preflightTemplateService) Detail(_ *gin.Context, id string) (*models.ScanTemplate, error) {
	f.ids = append(f.ids, id)
	return f.detail, f.err
}

type preflightNodeService struct {
	node.Service
	online []models.NodeData
	err    error
	flags  []bool
}

func (f *preflightNodeService) GetNodeData(_ context.Context, onlineOnly bool) ([]models.NodeData, error) {
	f.flags = append(f.flags, onlineOnly)
	if !onlineOnly {
		return append(append([]models.NodeData{}, f.online...), models.NodeData{Name: "offline-worker", State: "3"}), f.err
	}
	return f.online, f.err
}

type preflightCommonService struct {
	taskCommon.Service
	inserted []*models.Task
	err      error
}

func (f *preflightCommonService) Insert(_ context.Context, task *models.Task) (string, error) {
	f.inserted = append(f.inserted, task)
	return lifecycleTestID, f.err
}

func TestCreateScanTaskPreflight(t *testing.T) {
	for _, tc := range []struct {
		name   string
		modify func(*createScanTaskInput, *preflightTaskService, *preflightTemplateService, *preflightNodeService)
		want   string
	}{
		{"empty template", func(i *createScanTaskInput, _ *preflightTaskService, _ *preflightTemplateService, _ *preflightNodeService) {
			i.Template = ""
		}, "template"},
		{"invalid template", func(i *createScanTaskInput, _ *preflightTaskService, _ *preflightTemplateService, _ *preflightNodeService) {
			i.Template = "invalid"
		}, "template"},
		{"missing template", func(_ *createScanTaskInput, _ *preflightTaskService, f *preflightTemplateService, _ *preflightNodeService) {
			f.detail = nil
		}, "template not found"},
		{"template error", func(_ *createScanTaskInput, _ *preflightTaskService, f *preflightTemplateService, _ *preflightNodeService) {
			f.err = errors.New("template lookup unavailable")
		}, "template lookup unavailable"},
		{"no nodes", func(i *createScanTaskInput, _ *preflightTaskService, _ *preflightTemplateService, _ *preflightNodeService) {
			i.Node = nil
		}, "node"},
		{"empty node", func(i *createScanTaskInput, _ *preflightTaskService, _ *preflightTemplateService, _ *preflightNodeService) {
			i.Node = []string{""}
		}, "node"},
		{"blank node", func(i *createScanTaskInput, _ *preflightTaskService, _ *preflightTemplateService, _ *preflightNodeService) {
			i.Node = []string{"   "}
		}, "node"},
		{"offline node", func(i *createScanTaskInput, _ *preflightTaskService, _ *preflightTemplateService, _ *preflightNodeService) {
			i.Node = []string{"offline-worker"}
		}, "offline-worker"},
		{"unknown node", func(i *createScanTaskInput, _ *preflightTaskService, _ *preflightTemplateService, _ *preflightNodeService) {
			i.Node = []string{"unknown"}
		}, "unknown"},
		{"mixed nodes", func(i *createScanTaskInput, _ *preflightTaskService, _ *preflightTemplateService, _ *preflightNodeService) {
			i.Node = []string{"worker-a", "offline-worker"}
		}, "offline-worker"},
		{"node lookup error", func(_ *createScanTaskInput, _ *preflightTaskService, _ *preflightTemplateService, f *preflightNodeService) {
			f.err = errors.New("node lookup unavailable")
		}, "node lookup unavailable"},
		{"all nodes none online", func(i *createScanTaskInput, _ *preflightTaskService, _ *preflightTemplateService, f *preflightNodeService) {
			i.AllNode = true
			f.online = nil
		}, "online"},
		{"all nodes invalid explicit node", func(i *createScanTaskInput, _ *preflightTaskService, _ *preflightTemplateService, _ *preflightNodeService) {
			i.AllNode = true
			i.Node = []string{"offline-worker"}
		}, "offline-worker"},
		{"recurring", func(i *createScanTaskInput, _ *preflightTaskService, _ *preflightTemplateService, _ *preflightNodeService) {
			i.ScheduledTasks = true
			i.CycleType = "nhours"
			i.Hour = 1
		}, "scheduledTasks"},
		{"duplicate name", func(_ *createScanTaskInput, f *preflightTaskService, _ *preflightTemplateService, _ *preflightNodeService) {
			f.exists = true
		}, "already exists"},
		{"name lookup error", func(_ *createScanTaskInput, f *preflightTaskService, _ *preflightTemplateService, _ *preflightNodeService) {
			f.err = errors.New("name lookup unavailable")
		}, "name lookup unavailable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := createScanTaskInput{Name: "test", Template: lifecycleTestID, Node: []string{"worker-a"}, Target: "example.com"}
			tasks := &preflightTaskService{}
			templates := &preflightTemplateService{detail: &models.ScanTemplate{}}
			nodes := &preflightNodeService{online: []models.NodeData{{Name: "worker-a", State: "1"}}}
			common := &preflightCommonService{}
			tc.modify(&input, tasks, templates, nodes)
			useTestDeps(t, &deps{taskService: tasks, templateService: templates, nodeService: nodes, taskCommonService: common})
			result, _, err := createScanTask(context.Background(), nil, input)
			if err != nil {
				t.Fatal(err)
			}
			if !result.IsError || !strings.Contains(resultText(result), tc.want) || len(common.inserted) != 0 {
				t.Fatalf("result=%s inserted=%d", resultText(result), len(common.inserted))
			}
			if (tc.name == "empty template" || tc.name == "invalid template") && len(templates.ids) != 0 {
				t.Fatal("invalid template ID reached service")
			}
		})
	}
}

func TestCreateScanTaskPreservesConstruction(t *testing.T) {
	for _, source := range []string{"", "general", "project", "UrlScanSource"} {
		for _, allNode := range []bool{false, true} {
			t.Run(source+"/allNode="+map[bool]string{false: "false", true: "true"}[allNode], func(t *testing.T) {
				input := createScanTaskInput{Name: "test", Target: "example.com\nexample.org", Template: lifecycleTestID,
					Node: []string{"worker-b"}, AllNode: allNode, TargetSource: source, Project: []string{lifecycleTestID},
					BindProject: lifecycleTestID, Ignore: "ignored.example.com", Duplicates: "keep", TargetTp: "select",
					TargetIds: []string{lifecycleTestID}, Filter: map[string][]string{"project": {lifecycleTestID}}, Search: "test", TargetNumber: 10}
				if allNode {
					input.Node = nil
				}
				templates := &preflightTemplateService{detail: &models.ScanTemplate{}}
				nodes := &preflightNodeService{online: []models.NodeData{{Name: "worker-a", State: "1"}, {Name: "worker-b", State: "1"}}}
				common := &preflightCommonService{}
				useTestDeps(t, &deps{taskService: &preflightTaskService{}, templateService: templates, nodeService: nodes, taskCommonService: common})
				result, _, err := createScanTask(context.Background(), nil, input)
				if err != nil || result.IsError || len(common.inserted) != 1 {
					t.Fatalf("err=%v result=%s inserts=%d", err, resultText(result), len(common.inserted))
				}
				wantSource := source
				if source == "" {
					wantSource = "general"
				}
				wantNodes := []string{"worker-b"}
				if allNode {
					wantNodes = []string{"worker-a", "worker-b"}
				}
				want := &models.Task{Name: "test", Target: "example.com\nexample.org", Template: lifecycleTestID, Node: wantNodes,
					TargetSource: wantSource, Project: []string{lifecycleTestID}, BindProject: lifecycleTestID,
					Ignore: "ignored.example.com", Duplicates: "keep", TargetTp: "select", TargetIds: []string{lifecycleTestID},
					Filter: map[string][]interface{}{"project": {lifecycleTestID}}, Search: "test", TargetNumber: 10}
				if !reflect.DeepEqual(common.inserted[0], want) {
					t.Fatalf("inserted=%+v want=%+v", common.inserted[0], want)
				}
				if !reflect.DeepEqual(templates.ids, []string{lifecycleTestID}) || !reflect.DeepEqual(nodes.flags, []bool{true}) {
					t.Fatalf("template IDs=%v node flags=%v", templates.ids, nodes.flags)
				}
			})
		}
	}
}

func TestCreateScanTaskInsertError(t *testing.T) {
	common := &preflightCommonService{err: errors.New("insert unavailable")}
	useTestDeps(t, &deps{taskService: &preflightTaskService{}, templateService: &preflightTemplateService{detail: &models.ScanTemplate{}}, nodeService: &preflightNodeService{online: []models.NodeData{{Name: "worker-a", State: "1"}}}, taskCommonService: common})
	result, _, err := createScanTask(context.Background(), nil, createScanTaskInput{Name: "test", Template: lifecycleTestID, Node: []string{"worker-a"}, Target: "example.com"})
	if err != nil || !result.IsError || !strings.Contains(resultText(result), "insert unavailable") {
		t.Fatalf("err=%v result=%s", err, resultText(result))
	}
}

func TestCreateScanTaskAllOnlineNodesOverMCP(t *testing.T) {
	common := &preflightCommonService{}
	useTestDeps(t, &deps{taskService: &preflightTaskService{}, templateService: &preflightTemplateService{detail: &models.ScanTemplate{}}, nodeService: &preflightNodeService{online: []models.NodeData{{Name: "worker-a", State: "1"}}}, taskCommonService: common})
	ctx, session := testSession(t)
	result, err := session.CallTool(ctx, &mcpSDK.CallToolParams{Name: "create_scan_task", Arguments: map[string]any{
		"name": "test", "template": lifecycleTestID, "allNode": true, "target": "example.com\nexample.org",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError || len(common.inserted) != 1 || common.inserted[0].AllNode || !reflect.DeepEqual(common.inserted[0].Node, []string{"worker-a"}) {
		t.Fatalf("unexpected all-node request: %s inserts=%v", resultText(result), common.inserted)
	}
}

func TestCreateScanTaskRecurringRejectedBeforeDependencies(t *testing.T) {
	// No services: recurring requests must be rejected without any DB-capable call.
	useTestDeps(t, &deps{})
	ctx, session := testSession(t)
	result, err := session.CallTool(ctx, &mcpSDK.CallToolParams{Name: "create_scan_task", Arguments: map[string]any{
		"name": "test", "scheduledTasks": true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError || !strings.Contains(resultText(result), "one-shot") {
		t.Fatalf("unexpected result: %s", resultText(result))
	}
}
