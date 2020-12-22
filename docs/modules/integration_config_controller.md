# IntegrationConfig Controller

## What it does
- Creates `ServiceAccount/Secret` for git credentials
- Creates git webhook secret
- Registers webhook server for the git repository

## When it is called
- When `IntegrationConfig` is created/changes
