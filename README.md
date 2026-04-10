# taskflow-sabbavarapu

# TaskFlow Backend

Backend take-home assignment implementation for TaskFlow.

## Tech Stack
- Go
- PostgreSQL
- Docker
- JWT Authentication

## Project Status
Initial project setup completed. Core API implementation will be added in subsequent commits.

## Run locally

```bash
cp .env.example .env
set -a
source .env
set +a
cd backend
go run ./cmd/taskflow-sabbavarapu
```

Or with Docker:

```bash
docker compose up --build
```

## Notes
- `.env.example` is included for local setup
- actual `.env` should not be committed
- migrations, API routes, and auth will be added in later commits
