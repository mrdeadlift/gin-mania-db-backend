# Repository Guidelines

※本エージェントは英語で思考し、日本語で回答します。

## Project Structure & Module Organization
- `cmd/server/main.go`: HTTP entrypoint that wires routes and middleware.
- `internal/search`: Core search logic and in-memory catalogue data used for bootstrapping.
- `pkg/api`: Gin handlers, DTOs, and response helpers shared across routes.
- `pkg/database`: Database connection setup and Gorm helpers for repository layers.
- `db/migrations`: Versioned SQL migrations applied with `migrate`; keep files timestamped.
- `doc/`: Architecture, environment, and MVP notes—update when behavior or dependencies shift.

## Build, Test, and Development Commands
- `go mod tidy`: Sync module dependencies before builds and commits.
- `go run ./cmd/server`: Launch the API on `localhost:8080` for local development.
- `go build ./cmd/server`: Produce a deployable binary, surfacing compile issues early.
- `go test ./...`: Execute unit tests across `pkg/` and `internal/` packages.
- `docker-compose up postgres`: Start the local Postgres 15 instance (`gin-mania-postgres`).
- `migrate -path db/migrations -database "$DATABASE_URL" up`: Apply schema changes; pair with `down 1` when rolling back.

## Coding Style & Naming Conventions
- Target Go 1.22 semantics; format with `go fmt ./...` before committing.
- Use tabs for indentation and keep lines under 120 characters.
- Name packages with concise nouns (`search`, `database`) and files in lower_snake_case.
- Exported types/functions use PascalCase; unexported members stay camelCase.
- Prefer constructor functions in `pkg/` for injecting dependencies instead of global state.

## Testing Guidelines
- Place `_test.go` files beside implementation packages; follow table-driven patterns for handlers and search logic.
- Use the standard `testing` package; store fixtures under `testdata/` when required.
- Run `go test ./pkg/... ./internal/...` before pushing; aim for meaningful coverage (>70%) on new code.
- Capture regressions with focused tests whenever fixing bugs or adding migrations.

## Commit & Pull Request Guidelines
- Follow Conventional Commits (e.g., `feat(search): add fuzzy filter`); scope names mirror directories (`api`, `database`, `migrations`).
- Keep commits focused; run formatting, tests, and migration dry-runs prior to staging.
- Pull requests summarize behavior changes, list commands executed, document schema impacts, and attach curl examples or screenshots for new endpoints.
- Link related issues, request reviews from module owners, and refresh relevant docs under `doc/` when behavior shifts.
