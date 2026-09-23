# Agent Instructions

## Repository shape

- This is a Go 1.26 task HTTP microservice using Gin, PostgreSQL, and gRPC to the external `GoLearning-IdentityMicroService`.
- `main.go` is the composition root: it loads configuration, connects PostgreSQL, creates the identity client, repository, service, middleware, and routes.
- Keep dependency direction clear: HTTP handlers depend on service interfaces; the service depends on repository interfaces; PostgreSQL and identity integrations stay behind their package boundaries.
- Put domain types and domain errors in `internal/domain`; keep transport mapping in `internal/transport/http`.
- Define interfaces at the consuming layer and use Testify mocks for unit tests, following the existing tests in `tests/unit`.
- Preserve request context propagation because logging and identity middleware use `context.Context`.

## Commands

- Format changed Go files with `gofmt -w <files>`.
- Run unit tests with `go test ./tests/unit`.
- Run all package tests with `go test ./...`; integration tests require the external services described below.
- Run locally with `go run .` after providing the required environment variables.
- Build the service with `go build ./...`.
- Compose files live under `docker/`; invoke them explicitly from the repository root, for example `docker compose -f docker/docker-compose.yml up --build -d`.

## Environment and integration prerequisites

- Copy `.env.example` to a local `.env` and set `PORT`, `IDENTITY_SERVICE_ADDR`, and the `DB_*` variables. Never commit secrets or `.env` changes.
- Check the actual environment variable consumed by the code when changing configuration: the current `main.go` lookup uses `IDENTITY-SERVICE-ADDR`, while `.env.example` uses `IDENTITY_SERVICE_ADDR`; keep this mismatch visible and fix it deliberately rather than silently adding another spelling.
- `godotenv.Load("../.env")` is relative to the process working directory. Verify the working directory or environment variables before treating missing configuration as an application bug.
- Integration tests in `tests/integration` expect the task service at `localhost:8083` and the identity service at `localhost:8081`; both services must be running and seeded as required by their own repositories.
- Test Compose requires the external Docker network `todo-test-network` and an already available identity service. Production Compose similarly requires `todo-network`; neither Compose file starts the identity service.
- Database schema initialization comes from `docker/init-db.sql` when the PostgreSQL volume is first created. Recreate the volume when testing schema changes.

## HTTP contract

- Public health check: `GET /api/health`.
- Authenticated task routes are singular `/api/task`: `POST`, `GET`, `GET /:id`, `PUT /:id`, `DELETE /:id`, and `PATCH /complete/:id`.
- Route authentication is installed in `internal/transport/http/routes.go`; preserve middleware ordering when changing protected endpoints.
- Map domain errors through the existing HTTP mapper instead of inventing response shapes in individual handlers.

## Change and validation expectations

- Make focused changes and preserve existing public APIs unless the task requires a contract change.
- Add or update unit tests for handler/service behavior; use integration tests only when the change crosses the database, HTTP, or identity-service boundary.
- For database changes, update `docker/init-db.sql` and verify behavior against a fresh PostgreSQL volume.
- Run `gofmt` and the narrowest relevant test command before broader validation. Inspect `go test ./...` failures for missing external-service prerequisites before changing production code.
