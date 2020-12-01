# Scheduler

Scheduler is a module that schedules (actually creates) `PipelineRun`s for corresponding `IntegrationJob`s.

## Supported Schedulers
- FIFO & Max number of running `PipelineRun` restriction

## When it is called
- When `IntegrationJob` is created
- When `PipelineRun` is completed (whether or not there is an error)

## How it works
Procedure of the scheduler is as below.
1. `IntegrationJobController` initiates the scheduler by calling `scheduler.New()`
2. `scheduler.New()` creates a buffered channel with capacity 1, which acts as a 'schedule job' queue
3. `scheduler.New()` spawns a worker thread, listening the buffered channel
4. Whenever `IntegrationJob`s are created, or `PipelineRun`s are completed, `scheduler.Schedule()` is called by the `IntegrationJobController`
5. `scheduler.Schedule()` sends dummy int through buffered channel
6. (If worker thread is free) Worker thread calls an actual scheduling logic (only FIFO, for now)
7. (If worker thread is *not* free) Worker thread re-run the logic after it finishes
