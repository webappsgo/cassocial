package scheduler

import (
	"errors"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}
	if s.tasks == nil {
		t.Error("New() tasks map is nil")
	}
}

func TestScheduler_RegisterTask_Valid(t *testing.T) {
	s := New()
	err := s.RegisterTask("test-task", "* * * * * *", func() error { return nil })
	if err != nil {
		t.Errorf("RegisterTask() with valid schedule = %v, want nil", err)
	}
}

func TestScheduler_RegisterTask_InvalidSchedule(t *testing.T) {
	s := New()
	err := s.RegisterTask("bad-task", "not-a-schedule", func() error { return nil })
	if err == nil {
		t.Error("RegisterTask() with invalid schedule should return error")
	}
}

func TestScheduler_GetTaskStatus_Exists(t *testing.T) {
	s := New()
	_ = s.RegisterTask("mytask", "* * * * * *", func() error { return nil })

	task, err := s.GetTaskStatus("mytask")
	if err != nil {
		t.Errorf("GetTaskStatus() = %v, want nil", err)
	}
	if task == nil {
		t.Fatal("GetTaskStatus() returned nil task")
	}
	if task.Name != "mytask" {
		t.Errorf("task.Name = %q, want mytask", task.Name)
	}
	if !task.Enabled {
		t.Error("newly registered task should be enabled")
	}
}

func TestScheduler_GetTaskStatus_NotFound(t *testing.T) {
	s := New()
	_, err := s.GetTaskStatus("nonexistent")
	if err == nil {
		t.Error("GetTaskStatus() for nonexistent task should return error")
	}
}

func TestScheduler_ListTasks(t *testing.T) {
	s := New()
	_ = s.RegisterTask("task1", "* * * * * *", func() error { return nil })
	_ = s.RegisterTask("task2", "* * * * * *", func() error { return nil })

	tasks := s.ListTasks()
	if len(tasks) != 2 {
		t.Errorf("ListTasks() len = %d, want 2", len(tasks))
	}
}

func TestScheduler_EnableDisableTask(t *testing.T) {
	s := New()
	_ = s.RegisterTask("toggle", "* * * * * *", func() error { return nil })

	if err := s.DisableTask("toggle"); err != nil {
		t.Errorf("DisableTask() = %v, want nil", err)
	}
	task, _ := s.GetTaskStatus("toggle")
	if task.Enabled {
		t.Error("DisableTask() did not disable task")
	}

	if err := s.EnableTask("toggle"); err != nil {
		t.Errorf("EnableTask() = %v, want nil", err)
	}
	task, _ = s.GetTaskStatus("toggle")
	if !task.Enabled {
		t.Error("EnableTask() did not enable task")
	}
}

func TestScheduler_EnableTask_NotFound(t *testing.T) {
	s := New()
	if err := s.EnableTask("ghost"); err == nil {
		t.Error("EnableTask() on nonexistent task should return error")
	}
}

func TestScheduler_DisableTask_NotFound(t *testing.T) {
	s := New()
	if err := s.DisableTask("ghost"); err == nil {
		t.Error("DisableTask() on nonexistent task should return error")
	}
}

func TestScheduler_RunTaskNow_NotFound(t *testing.T) {
	s := New()
	if err := s.RunTaskNow("ghost"); err == nil {
		t.Error("RunTaskNow() on nonexistent task should return error")
	}
}

func TestScheduler_RunTaskNow_Success(t *testing.T) {
	s := New()
	executed := make(chan bool, 1)
	_ = s.RegisterTask("nowrun", "0 0 0 1 1 *", func() error {
		executed <- true
		return nil
	})

	if err := s.RunTaskNow("nowrun"); err != nil {
		t.Errorf("RunTaskNow() = %v, want nil", err)
	}

	select {
	case <-executed:
		// task ran
	case <-time.After(2 * time.Second):
		t.Error("RunTaskNow() task did not execute within timeout")
	}
}

func TestRunTask_TracksErrorCount(t *testing.T) {
	s := New()
	taskErr := errors.New("task failure")
	_ = s.RegisterTask("failing", "0 0 0 1 1 *", func() error { return taskErr })

	done := make(chan struct{})
	task, _ := s.GetTaskStatus("failing")
	// Override handler to signal when done
	origHandler := task.Handler
	task.Handler = func() error {
		err := origHandler()
		close(done)
		return err
	}

	go s.runTask(task)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runTask timed out")
	}

	if task.ErrorCount != 1 {
		t.Errorf("ErrorCount = %d, want 1", task.ErrorCount)
	}
	if task.LastError != taskErr.Error() {
		t.Errorf("LastError = %q, want %q", task.LastError, taskErr.Error())
	}
}

func TestRunTask_DisabledSkips(t *testing.T) {
	s := New()
	ran := false
	_ = s.RegisterTask("skip-disabled", "0 0 0 1 1 *", func() error {
		ran = true
		return nil
	})
	_ = s.DisableTask("skip-disabled")
	task, _ := s.GetTaskStatus("skip-disabled")

	s.runTask(task)

	if ran {
		t.Error("disabled task should not run")
	}
}

func TestScheduler_StartStop(t *testing.T) {
	s := New()
	_ = s.RegisterTask("heartbeat", "* * * * * *", func() error { return nil })
	s.Start()
	// Give it a moment to confirm it runs without panic
	time.Sleep(100 * time.Millisecond)
	s.Stop()
}
