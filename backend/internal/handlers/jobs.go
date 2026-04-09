package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// JobFunc is a function that a cron job executes.
type JobFunc func(ctx context.Context) (json.RawMessage, error)

// JobRegistry maps handler names to functions.
type JobRegistry struct {
	mu       sync.RWMutex
	handlers map[string]JobFunc
}

func NewJobRegistry() *JobRegistry {
	return &JobRegistry{handlers: make(map[string]JobFunc)}
}

func (r *JobRegistry) Register(name string, fn JobFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[name] = fn
}

func (r *JobRegistry) Get(name string) (JobFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.handlers[name]
	return fn, ok
}

func (r *JobRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.handlers))
	for name := range r.handlers {
		names = append(names, name)
	}
	return names
}

// JobsHandler manages cron job scheduling and execution.
type JobsHandler struct {
	jobs       domain.CronJobRepository
	executions domain.JobExecutionRepository
	registry   *JobRegistry
	scheduler  *cron.Cron
	entryMap   map[string]cron.EntryID // job ID -> cron entry ID
	mu         sync.Mutex
}

func NewJobsHandler(jobs domain.CronJobRepository, executions domain.JobExecutionRepository, registry *JobRegistry) *JobsHandler {
	return &JobsHandler{
		jobs:       jobs,
		executions: executions,
		registry:   registry,
		scheduler:  cron.New(cron.WithSeconds()),
		entryMap:   make(map[string]cron.EntryID),
	}
}

// StartScheduler loads all active jobs from the DB and starts the cron scheduler.
// Call this once on server boot.
func (h *JobsHandler) StartScheduler(ctx context.Context) error {
	jobs, err := h.jobs.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to load jobs: %w", err)
	}
	for _, job := range jobs {
		if !job.IsActive {
			continue
		}
		if err := h.scheduleJob(job); err != nil {
			// Log but don't abort — other jobs should still start.
			_ = err
		}
	}
	h.scheduler.Start()
	return nil
}

// scheduleJob adds a single job to the cron scheduler.
func (h *JobsHandler) scheduleJob(job domain.CronJob) error {
	fn, ok := h.registry.Get(job.Handler)
	if !ok {
		return fmt.Errorf("handler %q not registered", job.Handler)
	}

	jobID := job.ID
	entryID, err := h.scheduler.AddFunc(job.Schedule, func() {
		h.executeJob(context.Background(), jobID, fn)
	})
	if err != nil {
		return err
	}

	h.mu.Lock()
	h.entryMap[jobID] = entryID
	h.mu.Unlock()
	return nil
}

// executeJob runs a job and persists the execution record.
func (h *JobsHandler) executeJob(ctx context.Context, jobID string, fn JobFunc) {
	exec := &domain.JobExecution{
		ID:        uuid.NewString(),
		JobID:     jobID,
		Status:    "running",
		StartedAt: time.Now(),
	}
	_ = h.executions.Create(ctx, exec)

	output, err := fn(ctx)

	now := time.Now()
	duration := int(now.Sub(exec.StartedAt).Milliseconds())
	exec.FinishedAt = &now
	exec.DurationMs = &duration
	exec.Output = output

	if err != nil {
		exec.Status = "failed"
		errStr := err.Error()
		exec.Error = &errStr
	} else {
		exec.Status = "success"
	}

	_ = h.executions.Update(ctx, exec)
	_ = h.jobs.UpdateLastRun(ctx, jobID, now, nil)
}

// unscheduleJob removes a job from the scheduler by its job ID.
func (h *JobsHandler) unscheduleJob(jobID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if entryID, ok := h.entryMap[jobID]; ok {
		h.scheduler.Remove(entryID)
		delete(h.entryMap, jobID)
	}
}

// ---- HTTP handlers ----

// ListJobs godoc
// GET /admin/jobs
func (h *JobsHandler) ListJobs(c *fiber.Ctx) error {
	jobs, err := h.jobs.List(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch jobs"})
	}
	return c.JSON(fiber.Map{
		"jobs":     jobs,
		"handlers": h.registry.List(),
	})
}

// CreateJob godoc
// POST /admin/jobs
func (h *JobsHandler) CreateJob(c *fiber.Ctx) error {
	var req struct {
		Name     string `json:"name"`
		Schedule string `json:"schedule"`
		Handler  string `json:"handler"`
		IsActive bool   `json:"is_active"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Name == "" || req.Schedule == "" || req.Handler == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name, schedule, and handler are required"})
	}
	if _, ok := h.registry.Get(req.Handler); !ok {
		return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("handler %q is not registered", req.Handler)})
	}

	job := &domain.CronJob{
		ID:       uuid.NewString(),
		Name:     req.Name,
		Schedule: req.Schedule,
		Handler:  req.Handler,
		IsActive: req.IsActive,
	}
	if err := h.jobs.Create(c.Context(), job); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to create job"})
	}

	if req.IsActive {
		if err := h.scheduleJob(*job); err != nil {
			// Non-fatal: job is persisted, scheduling failed.
			return c.Status(201).JSON(fiber.Map{"job": job, "warning": err.Error()})
		}
	}
	return c.Status(201).JSON(fiber.Map{"job": job})
}

// UpdateJob godoc
// PUT /admin/jobs/:id
func (h *JobsHandler) UpdateJob(c *fiber.Ctx) error {
	id := c.Params("id")
	job, err := h.jobs.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "job not found"})
	}

	var req struct {
		Name     *string `json:"name"`
		Schedule *string `json:"schedule"`
		Handler  *string `json:"handler"`
		IsActive *bool   `json:"is_active"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Handler != nil {
		if _, ok := h.registry.Get(*req.Handler); !ok {
			return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("handler %q is not registered", *req.Handler)})
		}
		job.Handler = *req.Handler
	}

	reschedule := false
	if req.Name != nil {
		job.Name = *req.Name
	}
	if req.Schedule != nil {
		job.Schedule = *req.Schedule
		reschedule = true
	}
	if req.IsActive != nil {
		job.IsActive = *req.IsActive
		reschedule = true
	}

	if err := h.jobs.Update(c.Context(), job); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to update job"})
	}

	if reschedule {
		h.unscheduleJob(job.ID)
		if job.IsActive {
			if err := h.scheduleJob(*job); err != nil {
				return c.JSON(fiber.Map{"job": job, "warning": err.Error()})
			}
		}
	}
	return c.JSON(fiber.Map{"job": job})
}

// PauseJob godoc
// PUT /admin/jobs/:id/pause
func (h *JobsHandler) PauseJob(c *fiber.Ctx) error {
	id := c.Params("id")
	job, err := h.jobs.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "job not found"})
	}

	job.IsActive = false
	if err := h.jobs.Update(c.Context(), job); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to pause job"})
	}

	h.unscheduleJob(id)
	return c.JSON(fiber.Map{"job": job})
}

// ResumeJob godoc
// PUT /admin/jobs/:id/resume
func (h *JobsHandler) ResumeJob(c *fiber.Ctx) error {
	id := c.Params("id")
	job, err := h.jobs.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "job not found"})
	}

	job.IsActive = true
	if err := h.jobs.Update(c.Context(), job); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to resume job"})
	}

	if err := h.scheduleJob(*job); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("failed to schedule job: %s", err.Error())})
	}
	return c.JSON(fiber.Map{"job": job})
}

// RunNow godoc
// POST /admin/jobs/:id/run
func (h *JobsHandler) RunNow(c *fiber.Ctx) error {
	id := c.Params("id")
	job, err := h.jobs.FindByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "job not found"})
	}

	fn, ok := h.registry.Get(job.Handler)
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("handler %q is not registered", job.Handler)})
	}

	go h.executeJob(context.Background(), job.ID, fn)
	return c.JSON(fiber.Map{"message": "job triggered"})
}

// ListExecutions godoc
// GET /admin/jobs/:id/executions?page=1&limit=20
func (h *JobsHandler) ListExecutions(c *fiber.Ctx) error {
	id := c.Params("id")
	page, limit, offset := paginate(c)

	executions, total, err := h.executions.ListByJob(c.Context(), id, offset, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch executions"})
	}
	return c.JSON(fiber.Map{
		"executions": executions,
		"total":      total,
		"page":       page,
		"limit":      limit,
	})
}

// RetryExecution godoc
// POST /admin/jobs/:id/executions/:eid/retry
func (h *JobsHandler) RetryExecution(c *fiber.Ctx) error {
	jobID := c.Params("id")
	eid := c.Params("eid")

	_, err := h.executions.FindByID(c.Context(), eid)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "execution not found"})
	}

	job, err := h.jobs.FindByID(c.Context(), jobID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "job not found"})
	}

	fn, ok := h.registry.Get(job.Handler)
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("handler %q is not registered", job.Handler)})
	}

	go h.executeJob(context.Background(), job.ID, fn)
	return c.JSON(fiber.Map{"message": "retry triggered"})
}

// DeleteJob godoc
// DELETE /admin/jobs/:id
func (h *JobsHandler) DeleteJob(c *fiber.Ctx) error {
	id := c.Params("id")

	h.unscheduleJob(id)

	if err := h.jobs.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to delete job"})
	}
	return c.SendStatus(204)
}

// ListHandlers godoc
// GET /admin/jobs/handlers
func (h *JobsHandler) ListHandlers(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"handlers": h.registry.List()})
}
