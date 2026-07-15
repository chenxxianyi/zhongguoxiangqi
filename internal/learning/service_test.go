package learning

import (
	"testing"
	"time"

	"xiangqi-lab/internal/records"
)

func TestBuildActivateAndRollbackVersion(t *testing.T) {
	recordService := records.NewService()
	recordService.Import(records.ImportRequest{Format: "iccs", Content: "a3a4 a6a5", Result: "1-0"})
	service := NewService(recordService)
	firstJob := service.CreateJob(CreateJobRequest{Name: "v1"})
	first := waitJob(t, service, firstJob.ID)
	if first.Status != "completed" {
		t.Fatalf("job: %+v", first)
	}
	version, err := service.Activate(first.VersionID)
	if err != nil || version.Status != "active" {
		t.Fatalf("activate: %+v %v", version, err)
	}
	secondJob := service.CreateJob(CreateJobRequest{Name: "v2"})
	second := waitJob(t, service, secondJob.ID)
	if _, err := service.Activate(second.VersionID); err != nil {
		t.Fatal(err)
	}
	rolled, err := service.Rollback(first.VersionID)
	if err != nil || rolled.Status != "active" {
		t.Fatalf("rollback: %+v %v", rolled, err)
	}
}

func waitJob(t *testing.T, service *Service, id string) Job {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		job, err := service.GetJob(id)
		if err != nil {
			t.Fatal(err)
		}
		if job.Status == "completed" || job.Status == "failed" {
			return job
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("job timeout")
	return Job{}
}
