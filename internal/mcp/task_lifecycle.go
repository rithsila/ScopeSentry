package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type taskLifecycleInput struct {
	ID string `json:"id" jsonschema:"Single task MongoDB ObjectID"`
}

func stopScanTask(ctx context.Context, _ *mcp.CallToolRequest, input taskLifecycleInput) (*mcp.CallToolResult, any, error) {
	if _, err := primitive.ObjectIDFromHex(input.ID); err != nil {
		return errorResult("id must be a single valid task ObjectID", nil)
	}
	task, err := getDeps().taskService.GetTaskDetail(ginContext(ctx), input.ID)
	if err != nil {
		return errorResult("failed to get task; stop request was not sent", err)
	}
	if task == nil {
		return errorResult("task not found; use list_tasks to select an existing task", nil)
	}
	if task.Status == 3 || task.Progress >= 100 {
		return errorResult("task is completed and cannot be stopped", nil)
	}
	if err := getDeps().taskService.StopTasks(ginContext(ctx), []string{input.ID}); err != nil {
		return errorResult("failed to stop scan task", err)
	}
	return jsonToolResult(ginH{"success": true, "id": input.ID, "message": "stop request accepted; scanner acknowledgement is not guaranteed"})
}

func resumeScanTask(ctx context.Context, _ *mcp.CallToolRequest, input taskLifecycleInput) (*mcp.CallToolResult, any, error) {
	if _, err := primitive.ObjectIDFromHex(input.ID); err != nil {
		return errorResult("id must be a single valid task ObjectID", nil)
	}
	task, err := getDeps().taskService.GetTaskDetail(ginContext(ctx), input.ID)
	if err != nil {
		return errorResult("failed to get task; resume request was not sent", err)
	}
	if task == nil {
		return errorResult("task not found; use list_tasks to select an existing task", nil)
	}
	if task.Status == 3 || task.Progress >= 100 {
		return errorResult("task is completed and cannot be resumed", nil)
	}
	if task.Status != 2 {
		return errorResult("task must be paused before resuming; inspect get_task status", nil)
	}
	if err := getDeps().taskService.StartTasks(ginContext(ctx), []string{input.ID}); err != nil {
		return errorResult("failed to resume scan task", err)
	}
	return jsonToolResult(ginH{"success": true, "id": input.ID, "message": "resume request accepted; scanner acknowledgement is not guaranteed"})
}
