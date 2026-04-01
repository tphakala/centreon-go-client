# Contributing to centreon-go-client

## Development Setup

```bash
git clone https://github.com/tphakala/centreon-go-client.git
cd centreon-go-client
git config core.hooksPath .githooks
```

Requires Go 1.26+ and [golangci-lint](https://golangci-lint.run/) v2.11+.

## Making Changes

1. Create a feature branch from `main`
2. Write tests first (TDD)
3. Implement the feature
4. Run checks: `go test -race ./... && golangci-lint run ./...`
5. Commit (pre-commit hook runs automatically)
6. Open a PR against `main`

## Code Style

- Run `gofmt` — it's enforced by CI and pre-commit hook
- Follow existing patterns in the codebase
- One file per resource type (e.g., `hosts.go`, `host_groups.go`)
- Tests in `*_test.go` with the same package (`package centreon`)

## Adding a New Resource

1. Create `resource_name.go` with types and service struct
2. Create `resource_name_test.go` with tests using `newTestMux(t)`
3. Add service field to `Client` struct in `client.go`
4. Initialize in `NewClient()`

### Patterns

- **List-only resources**: See `commands.go`
- **Full CRUD (PUT)**: See `host_groups.go`
- **CRUD with PATCH**: See `hosts.go`
- **Read-only monitoring**: See `monitoring_hosts.go`

### Conventions

- `omitzero` on read-model optional fields
- `omitempty` on PATCH pointer fields
- `List()` + `All()` on every service
- `Get()` for direct GET endpoints, `GetByID()` for filtered list lookups
- `Create()` returns `(int, error)`
- `Update()` / `Delete()` return `error`

## Running Tests

```bash
# All tests
go test -race -count=1 ./...

# Specific test
go test -run TestHostService_Create -v ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Linting

```bash
golangci-lint run ./...
```

Uses 40+ linters including 12 custom ruleguard rules in `rules/`.
