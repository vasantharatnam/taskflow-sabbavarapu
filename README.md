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

## API Testing

Postman assets for quick backend review are available under:

- `backend/postman/TaskFlow.postman_collection.json`
- `backend/postman/TaskFlow.postman_environment.json`

Recommended request order:

1. Health Check
2. Login
3. Get Current User
4. List Projects
5. Create Project
6. Get Project By ID
7. Create Task
8. List Tasks By Project
9. Update Task
10. Delete Task
11. Delete Project

If you are using the current local `.env`, set the Postman `base_url` to `http://localhost:8081`.

## Notes
- `.env.example` is included for local setup
- actual `.env` should not be committed
- migrations, API routes, and auth will be added in later commits
