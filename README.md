# TaskFlow Backend

Backend take-home implementation for TaskFlow, a minimal task management API designed to demonstrate clean backend architecture, authentication, and relational data modeling.

## Tech Stack

- Go
- PostgreSQL
- `pgx` / `pgxpool`
- JWT authentication
- `bcrypt`
- `golang-migrate` with embedded SQL migrations
- Docker
- Docker Compose

## Features Implemented

- User registration and login
- JWT-based authentication with protected routes
- Current user endpoint via auth middleware
- Project CRUD with owner-based authorization
- Task CRUD with creator/owner-based authorization
- Task filtering by status and assignee
- Seeded user, project, and tasks for quick testing
- Automatic database migrations on application startup
- Postman collection for reviewer-friendly API walkthrough
- Clear separation of authentication (JWT) and authorization(resource-level access control)

## Architecture Decisions

- **Standard library HTTP with explicit routing**  
  I used `net/http` instead of a larger framework to keep the request flow easy to read during review. For a take-home assignment, explicit routing and middleware make the code easier to inspect and discuss.

- **Layered code by responsibility**  
  The backend is organized into `api`, `handlers`, `repository`, `models`, `auth`, `middleware`, `utils`, and `db`. HTTP request/response payloads live in `internal/api`, domain/database entities stay in `internal/models`, and repositories own SQL access. This keeps transport concerns separate from persistence models and makes handlers thinner.

- **Direct SQL via `pgx` instead of an ORM**  
  I chose `pgx` so the data access layer stays explicit and predictable. This helps with authorization-sensitive queries, migration clarity, and debugging, especially for a relational schema with straightforward CRUD behavior.

- **SQL migrations as the schema source of truth**  
  Schema changes live in versioned SQL files instead of auto-migration. That keeps database changes reviewable, deterministic, and aligned with the assignment requirement to manage schema via migrations.

- **Embedded migrations for zero-friction startup**  
  Migration files are embedded into the application and executed on startup. This reduces setup steps for reviewers and avoids a separate migration command in the common local flow.

- **Authorization kept close to resource access**  
  JWT validation is centralized in middleware, while ownership/access rules are enforced in handlers and repository queries where resource context is available. This makes route protection consistent while keeping authorization logic explicit.

- **Designed for clarity over premature optimization**  
  The system is intentionally simple and synchronous to keep the core API behavior easy to review. The structure allows scaling later (e.g., adding caching, background workers, or pagination) without major refactoring.

## Project Structure

```text
.
├── docker-compose.yml
├── .env.example
├── README.md
└── backend
    ├── cmd/taskflow-sabbavarapu
    │   └── main.go
    ├── internal
    │   ├── api
    │   ├── auth
    │   ├── config
    │   ├── db
    │   ├── handlers
    │   ├── middleware
    │   ├── models
    │   ├── repository
    │   ├── response
    │   └── utils
    ├── migrations
    └── postman
        ├── TaskFlow.postman_collection.json
        └── TaskFlow.postman_environment.json
```

## Running Locally

Assumption: Docker is installed.

1. Clone the repository
2. Copy the example environment file
3. Start the stack

```bash
git clone <your-repo-url>
cd taskflow-sabbavarapu
cp .env.example .env
docker compose up --build
```

If the API container exits on first startup, re-run `docker compose up` once Postgres is fully ready.

What this does:

- starts PostgreSQL
- builds and starts the Go API
- runs migrations automatically on startup
- loads seeded test data

Default local URLs with `.env.example`:

- API base URL: `http://localhost:8080`
- Health check: `http://localhost:8080/health`

If you change `APP_PORT` in `.env`, use the same port in Postman.

## Environment Variables

The project includes `.env.example` with sensible local defaults.

| Variable | Description | Default |
|---|---|---|
| `APP_ENV` | Application environment | `development` |
| `APP_PORT` | API port | `8080` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `postgres` |
| `DB_NAME` | PostgreSQL database name | `taskflow` |
| `DB_SSLMODE` | PostgreSQL SSL mode | `disable` |
| `JWT_SECRET` | JWT signing secret | required |
| `JWT_EXPIRATION_HOURS` | JWT expiry in hours | `24` |

## Test Credentials

Seeded credentials for immediate login:

```text
Email:    test@example.com
Password: password123
```

Seeded project ID:

```text
22222222-2222-2222-2222-222222222222
```

## API Overview

All endpoints (except `/auth/*` and `/health`) require:

Authorization: Bearer <token>

### Health

- `GET /health`

### Auth

- `POST /auth/register`
- `POST /auth/login`
- `GET /me`

### Projects

- `GET /projects`
- `POST /projects`
- `GET /projects/:id`
- `PATCH /projects/:id`
- `DELETE /projects/:id`

### Tasks

- `GET /projects/:id/tasks`
- `POST /projects/:id/tasks`
- `PATCH /tasks/:id`
- `DELETE /tasks/:id`

### Minimal Examples

Register:

```http
POST /auth/register
Content-Type: application/json
```

```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "secret123"
}
```

Login:

```http
POST /auth/login
Content-Type: application/json
```

```json
{
  "email": "test@example.com",
  "password": "password123"
}
```

Create project:

```http
POST /projects
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "name": "My Project",
  "description": "Project created for review"
}
```

Create task:

```http
POST /projects/:id/tasks
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "title": "Finish backend API",
  "description": "Implement required endpoints",
  "status": "todo",
  "priority": "high",
  "due_date": "2026-04-30"
}
```

## Postman Collection Usage

Postman assets are included for fast API review:

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

The collection stores:

- `token`
- `user_id`
- `project_id`
- `task_id`

so reviewers can exercise the full flow without copying values manually.

Important:

- Set `base_url` to match `APP_PORT`
- `.env.example` uses `http://localhost:8080`
- if your local `.env` uses a different port, update the Postman environment accordingly

## Error Response Format

Validation errors:

```json
{
  "error": "validation failed",
  "fields": {
    "email": "is required"
  }
}
```

Unauthorized:

```json
{
  "error": "unauthorized"
}
```

Forbidden:

```json
{
  "error": "forbidden"
}
```

Not found:

```json
{
  "error": "not found"
}
```

## Assumptions / Tradeoffs

- I prioritized a clear, reviewable backend over adding extra infrastructure or framework abstraction.
- I used direct SQL with `pgx` instead of an ORM so queries and authorization-sensitive data access stay explicit.
- Automatic embedded migrations improve reviewer experience, but they trade off some operational flexibility compared with a dedicated migration command in production systems.
- I focused on the assignment’s core API behavior first; optional features such as pagination, stats endpoints, or wider automated test coverage were left secondary to a working end-to-end backend.
- The service is intentionally simple and synchronous: no background workers, caching layer, or event-driven components were introduced because they would add complexity without improving the assignment’s core review goals.
- Authorization decisions are enforced explicitly rather than abstracted away, so they remain easy to audit and reason about during review.

## Improvements

With more time, I would add:

- broader integration test coverage for auth, project, and task flows
- pagination on list endpoints
- a `GET /projects/:id/stats` endpoint
- request IDs and richer structured logging
- stronger request validation and clearer domain-level error mapping
- CI for tests, linting, and container validation
- OpenAPI documentation alongside the Postman collection
- Introduce service layer abstraction if business logic grows beyond simple CRUD

## Submission Notes

- This is a backend-only submission.
- Docker Compose is the primary local setup path.
- Migrations run automatically on startup.
- Seed data is included for immediate testing.
- A Postman collection is included so the API can be reviewed quickly without manually crafting requests.
