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
- Identity configuration is read from `IDENTITY_SERVICE_ADDR`; the hyphenated `IDENTITY-SERVICE-ADDR` spelling remains a compatibility fallback.
- `godotenv.Load("../.env")` is relative to the process working directory. Verify the working directory or environment variables before treating missing configuration as an application bug.
- Integration tests in `tests/integration` expect the task service at `localhost:8083` and the identity service at `localhost:8081`; both services must be running and seeded as required by their own repositories.
- Test Compose requires the external Docker network `todo-test-network` and an already available identity service. Production Compose similarly requires `todo-network`; neither Compose file starts the identity service.
- Database schema changes are applied by embedded `golang-migrate` migrations during service startup. Use `make migrate-up` for explicit migration runs.

## HTTP contract

- Public health check: `GET /api/health`.
- Authenticated task routes are singular `/api/task`: `POST`, `GET`, `GET /:id`, `PUT /:id`, `DELETE /:id`, and `PATCH /complete/:id`.
- Route authentication is installed in `internal/transport/http/routes.go`; preserve middleware ordering when changing protected endpoints.
- Map domain errors through the existing HTTP mapper instead of inventing response shapes in individual handlers.

## Change and validation expectations

- Make focused changes and preserve existing public APIs unless the task requires a contract change.
- Add or update unit tests for handler/service behavior; use integration tests only when the change crosses the database, HTTP, or identity-service boundary.
- For database changes, add ordered up/down migrations under `internal/store/migrations` and verify behavior against an empty database and the previous migration version.
- Run `gofmt` and the narrowest relevant test command before broader validation. Inspect `go test ./...` failures for missing external-service prerequisites before changing production code.

## Gin-specific guidance

- Use `gin.New()` instead of `gin.Default()` and attach only the middleware needed for the service.
- Keep handlers thin: parse input, call a service, marshal output; keep business logic in the service layer.
- Define request and response structs per endpoint; do not reuse domain model structs as API DTOs.
- Register middleware as factory functions and preserve the ordering of global and route-level middleware.
- Abort middleware early with `c.AbortWithStatusJSON(...)` when authentication or validation fails; do not continue after aborting.
- Use `c.Set(...)` and `c.MustGet(...)` for request-scoped values such as user or request identifiers.
- Use `c.ShouldBindJSON`, `c.ShouldBindQuery`, and `c.ShouldBindUri` for request binding and validate business rules in the service layer.
- Centralize HTTP error mapping through the existing mapper and avoid returning ad hoc JSON shapes in handlers.
- Propagate request context through `c.Request.Context()` when calling downstream dependencies so cancellation and logging context remain intact.
- Use Gin security defaults appropriate to production, including release mode and security headers, but keep the repo's existing middleware layout and route structure intact.
- Unit tests should use `httptest.NewRecorder()` with `router.ServeHTTP(...)` and assert both the HTTP status and the JSON body for the real handler behavior.
