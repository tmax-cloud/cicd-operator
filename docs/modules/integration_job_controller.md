# IntegrationJob Controller

## What it does
- Triggers `scheduler.Schedule` function when `IntegrationJob` is created
- Triggers `shceduler.Schedule` function when corresponding `PipelineRun` is completed
- Updates `IntegrationJob`'s status referring to corresponding `PipelineRun`

## When it is called
- Whenever changes occur to `IntegrationJob` / `PipelineRun`
