package task

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Autumn-27/ScopeSentry/internal/models"
	repoTask "github.com/Autumn-27/ScopeSentry/internal/repositories/task/task"
	"github.com/Autumn-27/ScopeSentry/internal/services/node"
	taskCommon "github.com/Autumn-27/ScopeSentry/internal/services/task/common"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const lifecycleID = "507f1f77bcf86cd799439011"

type lifecycleRepo struct {
	repoTask.Repository
	task               *models.Task
	findErr, updateErr error
	updates            int
	filter, update     bson.M
	lookupID           primitive.ObjectID
}

func (f *lifecycleRepo) FindByID(_ context.Context, id primitive.ObjectID) (*models.Task, error) {
	f.lookupID = id
	return f.task, f.findErr
}
func (f *lifecycleRepo) UpdateOne(_ context.Context, filter bson.M, update bson.M) error {
	f.updates++
	f.filter, f.update = filter, update
	return f.updateErr
}

type lifecycleNode struct {
	node.Service
	err      error
	messages []models.Message
}

func (f *lifecycleNode) RefreshConfig(_ context.Context, message models.Message) error {
	f.messages = append(f.messages, message)
	return f.err
}

type lifecycleCommon struct {
	taskCommon.Service
	err         error
	ids         []string
	resumeFlags []bool
}

func (f *lifecycleCommon) CreateTaskScan(_ context.Context, _ models.Task, id string, resume bool) (int64, error) {
	f.ids = append(f.ids, id)
	f.resumeFlags = append(f.resumeFlags, resume)
	return 0, f.err
}

func TestStopTasksPropagatesErrors(t *testing.T) {
	for _, tc := range []struct {
		name               string
		nodeErr, updateErr error
		want               string
		wantUpdates        int
	}{
		{"broadcast", errors.New("redis unavailable"), nil, "redis unavailable", 0},
		{"update", nil, errors.New("mongo unavailable"), "mongo unavailable", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &lifecycleRepo{updateErr: tc.updateErr}
			nodes := &lifecycleNode{err: tc.nodeErr}
			s := &service{taskRepo: repo, nodeService: nodes}
			err := s.StopTasks(&gin.Context{}, []string{lifecycleID})
			if err == nil || !strings.Contains(err.Error(), tc.want) || repo.updates != tc.wantUpdates {
				t.Fatalf("err=%v updates=%d", err, repo.updates)
			}
		})
	}
}

func TestStartTasksPropagatesErrors(t *testing.T) {
	for _, tc := range []struct {
		name                        string
		findErr, scanErr, updateErr error
		missing                     bool
		want                        string
		scans, updates              int
	}{
		{"lookup", errors.New("lookup unavailable"), nil, nil, false, "lookup unavailable", 0, 0},
		{"missing", nil, nil, nil, true, "not found", 0, 0},
		{"dispatch", nil, errors.New("queue unavailable"), nil, false, "queue unavailable", 1, 0},
		{"update", nil, nil, errors.New("update unavailable"), false, "update unavailable", 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &lifecycleRepo{task: &models.Task{Status: 2}, findErr: tc.findErr, updateErr: tc.updateErr}
			if tc.missing {
				repo.task = nil
			}
			common := &lifecycleCommon{err: tc.scanErr}
			s := &service{taskRepo: repo, commonService: common}
			err := s.StartTasks(&gin.Context{}, []string{lifecycleID})
			if err == nil || !strings.Contains(err.Error(), tc.want) || len(common.ids) != tc.scans || repo.updates != tc.updates {
				t.Fatalf("err=%v scans=%d updates=%d", err, len(common.ids), repo.updates)
			}
		})
	}
}

func TestLifecycleRejectsInvalidIDs(t *testing.T) {
	for _, id := range []string{"", "bad", lifecycleID + "," + lifecycleID} {
		for _, operation := range []string{"stop", "resume"} {
			t.Run(operation+"/"+id, func(t *testing.T) {
				s := &service{taskRepo: &lifecycleRepo{}, nodeService: &lifecycleNode{}}
				var err error
				if operation == "stop" {
					err = s.StopTasks(&gin.Context{}, []string{id})
				} else {
					err = s.StartTasks(&gin.Context{}, []string{id})
				}
				if err == nil || !strings.Contains(err.Error(), "invalid task id") {
					t.Fatalf("err=%v", err)
				}
			})
		}
	}
}

func TestLifecyclePersistsStatus(t *testing.T) {
	for _, op := range []string{"stop", "resume"} {
		t.Run(op, func(t *testing.T) {
			repo := &lifecycleRepo{task: &models.Task{Status: 2}}
			nodes, common := &lifecycleNode{}, &lifecycleCommon{}
			s := &service{taskRepo: repo, nodeService: nodes, commonService: common}
			var err error
			status := 2
			if op == "stop" {
				err = s.StopTasks(&gin.Context{}, []string{lifecycleID})
			} else {
				status = 1
				err = s.StartTasks(&gin.Context{}, []string{lifecycleID})
			}
			if err != nil {
				t.Fatal(err)
			}
			id, _ := primitive.ObjectIDFromHex(lifecycleID)
			if repo.updates != 1 || repo.update["$set"].(bson.M)["status"] != status || repo.filter["_id"] != id {
				t.Fatalf("filter=%v update=%v", repo.filter, repo.update)
			}
			if op == "stop" && (len(nodes.messages) != 1 || nodes.messages[0] != (models.Message{Name: "all", Type: "stop_task", Content: lifecycleID})) {
				t.Fatalf("messages=%v", nodes.messages)
			}
			if op == "resume" && (len(common.ids) != 1 || common.ids[0] != lifecycleID || !common.resumeFlags[0]) {
				t.Fatalf("resume ids=%v flags=%v", common.ids, common.resumeFlags)
			}
		})
	}
}

func TestLifecycleCanonicalizesIDs(t *testing.T) {
	for _, id := range []string{"507F1F77BCF86CD799439011", "507f1F77bCf86cD799439011"} {
		for _, op := range []string{"stop", "resume"} {
			t.Run(op+"/"+id, func(t *testing.T) {
				repo := &lifecycleRepo{task: &models.Task{Status: 2}}
				nodes, common := &lifecycleNode{}, &lifecycleCommon{}
				s := &service{taskRepo: repo, nodeService: nodes, commonService: common}
				var err error
				status := 2
				if op == "stop" {
					err = s.StopTasks(&gin.Context{}, []string{id})
				} else {
					status = 1
					err = s.StartTasks(&gin.Context{}, []string{id})
				}
				if err != nil {
					t.Fatal(err)
				}
				oid, _ := primitive.ObjectIDFromHex(lifecycleID)
				if repo.updates != 1 || repo.filter["_id"] != oid || repo.update["$set"].(bson.M)["status"] != status {
					t.Fatalf("filter=%v update=%v", repo.filter, repo.update)
				}
				if op == "stop" && (len(nodes.messages) != 1 || nodes.messages[0] != (models.Message{Name: "all", Type: "stop_task", Content: lifecycleID})) {
					t.Fatalf("stop broadcast must use canonical lowercase id: %v", nodes.messages)
				}
				if op == "resume" && (repo.lookupID != oid || len(common.ids) != 1 || common.ids[0] != lifecycleID || !common.resumeFlags[0]) {
					t.Fatalf("resume lookup=%v ids=%v flags=%v", repo.lookupID, common.ids, common.resumeFlags)
				}
			})
		}
	}
}

func TestStartTasksRejectsNonPausedOrCompleted(t *testing.T) {
	for _, tc := range []struct {
		name string
		task models.Task
		want string
	}{
		{"running", models.Task{Status: 1}, "paused"},
		{"completed status", models.Task{Status: 3}, "completed"},
		{"completed progress", models.Task{Status: 2, Progress: 100}, "completed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &lifecycleRepo{task: &tc.task}
			common := &lifecycleCommon{}
			s := &service{taskRepo: repo, commonService: common}
			err := s.StartTasks(&gin.Context{}, []string{lifecycleID})
			if err == nil || !strings.Contains(err.Error(), tc.want) || len(common.ids) != 0 || repo.updates != 0 {
				t.Fatalf("err=%v dispatches=%v updates=%d", err, common.ids, repo.updates)
			}
		})
	}
}
