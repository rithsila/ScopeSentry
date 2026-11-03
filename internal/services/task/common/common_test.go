package common

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Autumn-27/ScopeSentry/internal/models"
	"github.com/Autumn-27/ScopeSentry/internal/repositories/task/task"
	"github.com/Autumn-27/ScopeSentry/internal/services/assets/url"
	"github.com/Autumn-27/ScopeSentry/internal/services/dictionary"
)

type missingTemplateRepo struct{ task.Repository }

func (*missingTemplateRepo) FindTemplateByID(_ context.Context, _ string) (*models.ScanTemplate, error) {
	return nil, nil
}

func TestGetScanTemplateMissing(t *testing.T) {
	s := &service{taskRepo: &missingTemplateRepo{}}
	defer func() {
		if p := recover(); p != nil {
			t.Errorf("missing template must return error, got panic: %v", p)
		}
	}()
	template, err := s.GetScanTemplate(context.Background(), &models.Task{Template: "507f1f77bcf86cd799439011"})
	if err == nil || !strings.Contains(err.Error(), "template not found") || template != nil {
		t.Fatalf("template=%v err=%v", template, err)
	}
}

type dispatchRepo struct {
	task.Repository
	nodes []string
	err   error
}

func (*dispatchRepo) FindTemplateByID(_ context.Context, _ string) (*models.ScanTemplate, error) {
	return &models.ScanTemplate{}, nil
}
func (f *dispatchRepo) RPushNodeTask(_ context.Context, name string, _ any) error {
	f.nodes = append(f.nodes, name)
	if name == "worker-a" {
		return f.err
	}
	return nil
}

type emptyDictionaries struct{ dictionary.ManageService }

func (*emptyDictionaries) List(context.Context) ([]models.DictionaryMeta, error) { return nil, nil }

type emptyPorts struct{ dictionary.PortService }

func (*emptyPorts) Get(context.Context, string, int, int) ([]models.PortDoc, int64, error) {
	return nil, 0, nil
}

func TestCreateTaskScanReportsDispatchFailure(t *testing.T) {
	repo := &dispatchRepo{err: errors.New("queue unavailable")}
	s := &service{taskRepo: repo, dictionaryService: &emptyDictionaries{}, portsService: &emptyPorts{}}
	_, err := s.CreateTaskScan(context.Background(), models.Task{Node: []string{"worker-a", "worker-b"}}, "507f1f77bcf86cd799439011", true)
	if err == nil || !strings.Contains(err.Error(), "queue unavailable") || !strings.Contains(err.Error(), "worker-a") {
		t.Fatalf("dispatch error lost: %v", err)
	}
	if len(repo.nodes) != 2 || repo.nodes[1] != "worker-b" {
		t.Fatalf("remaining nodes not attempted: %v", repo.nodes)
	}
}

func TestInsertURLTargets(t *testing.T) {
	want := errors.New("URL source unavailable")
	urls := &urlTargets{err: want}
	// Replace the external URL lookup, not all repositories via the real constructor.
	s := &service{urlService: urls}
	_, err := s.Insert(context.Background(), &models.Task{TargetSource: "UrlScanSource", TargetTp: "select", TargetIds: []string{"507f1f77bcf86cd799439011"}})
	if !errors.Is(err, want) || len(urls.ids) != 1 || urls.ids[0] != "507f1f77bcf86cd799439011" {
		t.Fatalf("URL dependency not used: ids=%v err=%v", urls.ids, err)
	}
}

type urlTargets struct {
	url.Service
	ids []string
	err error
}

func (f *urlTargets) GetTaskTargetByIDs(_ context.Context, ids []string) (string, error) {
	f.ids = ids
	return "", f.err
}
