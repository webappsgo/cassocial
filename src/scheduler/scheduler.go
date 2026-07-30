package scheduler

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Scheduler manages background tasks using the built-in cron implementation.
// It requires no external scheduler (PART 19).
type Scheduler struct {
	tasks  map[string]*Task
	mu     sync.RWMutex
	logger *log.Logger

	stop    chan struct{}
	wg      sync.WaitGroup
	running bool
}

// Task represents a scheduled task
type Task struct {
	Name       string
	Schedule   string // Cron expression (6 fields, with seconds)
	Handler    func() error
	LastRun    time.Time
	NextRun    time.Time
	RunCount   int
	ErrorCount int
	LastError  string
	Enabled    bool

	sched     *cronSchedule
	lastFired time.Time
}

// New creates a new scheduler
func New() *Scheduler {
	return &Scheduler{
		tasks:  make(map[string]*Task),
		logger: log.New(log.Writer(), "[scheduler] ", log.LstdFlags),
	}
}

// Start starts the scheduler tick loop
func (s *Scheduler) Start() {
	s.logger.Println("Starting scheduler...")

	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stop = make(chan struct{})
	s.mu.Unlock()

	s.wg.Add(1)
	go s.loop()

	s.logger.Println("Scheduler started")
}

// Stop stops the scheduler tick loop
func (s *Scheduler) Stop() {
	s.logger.Println("Stopping scheduler...")

	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stop)
	s.mu.Unlock()

	s.wg.Wait()
	s.logger.Println("Scheduler stopped")
}

// loop ticks once per second and dispatches any tasks whose schedule matches.
func (s *Scheduler) loop() {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case now := <-ticker.C:
			s.tick(now)
		}
	}
}

// tick dispatches all tasks whose schedule matches the truncated current second.
func (s *Scheduler) tick(now time.Time) {
	now = now.Truncate(time.Second)

	s.mu.Lock()
	var due []*Task
	for _, task := range s.tasks {
		if task.sched == nil || !task.sched.match(now) {
			continue
		}
		// Guard against dispatching the same task twice within one second.
		if task.lastFired.Equal(now) {
			continue
		}
		task.lastFired = now
		task.NextRun = task.sched.next(now)
		due = append(due, task)
	}
	s.mu.Unlock()

	for _, task := range due {
		go s.runTask(task)
	}
}

// RegisterTask registers a new task
func (s *Scheduler) RegisterTask(name, schedule string, handler func() error) error {
	sched, err := parseCron(schedule)
	if err != nil {
		return fmt.Errorf("failed to register task %s: %w", name, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	task := &Task{
		Name:     name,
		Schedule: schedule,
		Handler:  handler,
		Enabled:  true,
		sched:    sched,
		NextRun:  sched.next(time.Now()),
	}

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
