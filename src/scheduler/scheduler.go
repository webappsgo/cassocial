package scheduler

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Scheduler manages background tasks and cron jobs
type Scheduler struct {
	cron   *cron.Cron
	tasks  map[string]*Task
	mu     sync.RWMutex
	logger *log.Logger
}

// Task represents a scheduled task
type Task struct {
	Name        string
	Schedule    string // Cron expression
	Handler     func() error
	LastRun     time.Time
	NextRun     time.Time
	RunCount    int
	ErrorCount  int
	LastError   string
	Enabled     bool
}

// New creates a new scheduler
func New() *Scheduler {
	return &Scheduler{
		cron:   cron.New(cron.WithSeconds()),
		tasks:  make(map[string]*Task),
		logger: log.New(log.Writer(), "[scheduler] ", log.LstdFlags),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.logger.Println("Starting scheduler...")
	s.cron.Start()
	s.logger.Println("Scheduler started")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.logger.Println("Stopping scheduler...")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Println("Scheduler stopped")
}

// RegisterTask registers a new task
func (s *Scheduler) RegisterTask(name, schedule string, handler func() error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create task
	task := &Task{
		Name:     name,
		Schedule: schedule,
		Handler:  handler,
		Enabled:  true,
	}

	// Wrap handler to track execution
	wrappedHandler := func() {
		s.runTask(task)
	}

	// Add to cron
	entryID, err := s.cron.AddFunc(schedule, wrappedHandler)
	if err != nil {
		return fmt.Errorf("failed to register task %s: %w", name, err)
	}

	// Get next run time
	entry := s.cron.Entry(entryID)
	task.NextRun = entry.Next

	s.tasks[name] = task
	s.logger.Printf("Registered task: %s (schedule: %s, next run: %s)", name, schedule, task.NextRun)

	return nil
}

// runTask executes a task and tracks its status
func (s *Scheduler) runTask(task *Task) {
	if !task.Enabled {
		return
	}

	s.logger.Printf("Running task: %s", task.Name)
	start := time.Now()

	s.mu.Lock()
	task.LastRun = start
	task.RunCount++
	s.mu.Unlock()

	// Execute task
	if err := task.Handler(); err != nil {
		s.mu.Lock()
		task.ErrorCount++
		task.LastError = err.Error()
		s.mu.Unlock()

		s.logger.Printf("Task %s failed: %v", task.Name, err)
	} else {
		s.mu.Lock()
		task.LastError = ""
		s.mu.Unlock()

		duration := time.Since(start)
		s.logger.Printf("Task %s completed in %v", task.Name, duration)
	}
}

// GetTaskStatus returns the status of a task
func (s *Scheduler) GetTaskStatus(name string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[name]
	if !exists {
		return nil, fmt.Errorf("task %s not found", name)
	}

	return task, nil
}

// ListTasks returns all registered tasks
func (s *Scheduler) ListTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// EnableTask enables a task
func (s *Scheduler) EnableTask(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[name]
	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	task.Enabled = true
	s.logger.Printf("Task %s enabled", name)
	return nil
}

// DisableTask disables a task
func (s *Scheduler) DisableTask(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[name]
	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	task.Enabled = false
	s.logger.Printf("Task %s disabled", name)
	return nil
}

// RunTaskNow runs a task immediately (outside its schedule)
func (s *Scheduler) RunTaskNow(name string) error {
	s.mu.RLock()
	task, exists := s.tasks[name]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("task %s not found", name)
	}

	s.logger.Printf("Running task on demand: %s", name)
	go s.runTask(task)

	return nil
}
