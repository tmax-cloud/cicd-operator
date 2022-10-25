# Scheduler

Scheduler is a module that schedules (actually creates) `PipelineRun`s for corresponding `IntegrationJob`s.

## Supported Schedulers
- FIFO & Max number of running `PipelineRun` restriction

## When it is called
- When `IntegrationJob` is created/deleted/updated
- When `PipelineRun` is completed/deleted

## How it works
Procedure of the scheduler is as below.
1. `main.go` initiates the scheduler by calling `scheduler.New()`
2. `scheduler.New()` creates a `JobPool`, which is a local cache for running/pending `IntegrationJob`s
3. Whenever `IntegrationJob`s' status changes, `scheduler.Notify()` is called
4. `scheduler.Notify()` calls `jobPool.SyncJob()` so that the local cache is synced to the current state
5. Whenever `IntegrationJob`s are created, or `PipelineRun`s are completed, `jobPool` calls `sendSchedule()` to send a signal through a buffered channel to schedule a new `PipelineRun`
6. (If worker thread is free) Worker thread calls an actual scheduling logic (only FIFO, for now)
7. (If worker thread is *not* free) Worker thread re-run the logic after it finishes
