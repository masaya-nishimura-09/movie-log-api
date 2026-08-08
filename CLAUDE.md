# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

A Go REST API for a movie log app, currently implementing user registration/profile management and JWT-based auth (login/logout/refresh). Built with Gin + GORM/Postgres.

## Commands

```bash
# Run the API locally (loads .env, requires postgres reachable via POSTGRES_DSN)
go run ./cmd/movie-log-api

# Build
go build ./...

# Start Postgres (creates schema from scripts/create_tables.sql, seeds from scripts/seed.sql)
docker compose up -d

# Run all tests
go test ./...

# Run a single test
go test ./internal/domain/user -run TestValidateUsername

# Vet
go vet ./...
```

There is no lint config, CI workflow, or Makefile in this repo — the commands above are the whole toolchain.

Required env vars (see `.env.example`): `POSTGRES_DSN`, `JWT_SECRET`, plus optional `ACCESS_TOKEN_TTL_HOURS` (default 24h) and `REFRESH_TOKEN_TTL_HOURS` (default 30d). `.env` is loaded via `godotenv` in `main.go` and is gitignored.

**Known issue:** `go vet ./...` / `go test ./...` currently fail to build because `internal/domain/user/user_test.go` calls `ValidateEmail`/`ValidatePassword`, which don't exist — the real functions are `NewEmail`/`NewPassword` (see `email.go`/`password.go`). `internal/infrastructure/auth/refresh_token_repository_test.go` is also just a stub (`package auth`, no tests). Don't assume a red `go test ./...` means you broke something — check whether it's this pre-existing issue first.

## Architecture

The codebase follows a strict layered (clean/onion) architecture, split by bounded context (`auth`, `user`) that cuts across every layer. When adding a feature, expect to touch one file in each of these layers, in this dependency order:

```
domain  →  usecase  →  handler
   ↑
infrastructure (implements domain interfaces)
```

- **`internal/domain/<context>/`** — pure business types and interface contracts, no framework or DB imports. Value objects (e.g. `user.Email`, `user.Password`, `user.Username`) are constructed via `NewX(value string) (X, error)` constructors that self-validate and return `exception.ErrInvalid`-wrapped errors — never construct these types by casting a raw string outside of tests. Repository/service contracts (`user.UserRepository`, `auth.AccessTokenService`, `auth.RefreshTokenRepository`) are defined here as interfaces; only their implementations live in `infrastructure`.
- **`internal/domain/exception/`** — the sentinel errors used across all layers (`ErrNotFound`, `ErrAlreadyExists`, `ErrInvalid`, `ErrUnauthenticated`). Lower layers return these wrapped with `fmt.Errorf("...: %w", exception.ErrX)`; handlers unwrap with `errors.Is` to pick an HTTP status. Stick to this vocabulary rather than inventing new sentinel errors.
- **`internal/usecase/<context>/`** — application services that orchestrate one or more repositories/domain services per operation (e.g. `AuthUsecase.Login` looks up the user, verifies bcrypt password, then issues both an access token and a refresh token). Takes domain interfaces as constructor dependencies, never a concrete infrastructure type — this is what makes handlers/usecases testable against fakes.
- **`internal/infrastructure/<context>/`** — GORM/JWT/etc. implementations of the domain interfaces. Each repository defines its own private `xDTO` struct (with `TableName()`) separate from the domain struct, and maps between them explicitly (`toDTO`/manual field mapping) rather than using the domain struct as the GORM model directly.
- **`internal/handler/<context>/`** — Gin handlers. Each parses/binds the request DTO, converts primitive fields to domain value objects via the `NewX` constructors (surfacing validation errors as `400 INVALID_INPUT`), calls the usecase, then maps `exception` sentinels to HTTP status codes. Every JSON error response has the shape `{"code": "...", "message": "..."}` — match this convention for new endpoints instead of returning bare error strings.
- **`internal/middleware/jwt.go`** — `JWTAuth` gin middleware: validates the `Authorization: Bearer <token>` header via `AuthUsecase.ValidateAccessToken`, re-checks the user still exists via `UserUsecase.GetByID`, then sets `userID`/`role` on the gin context for downstream handlers (see `getUserID` helper in `handler/user/handler.go`) to read.
- **`cmd/movie-log-api/main.go`** — the only place wiring happens: config loading → infrastructure → usecase → handler → gin router/route groups. This is manual dependency injection; there is no DI framework or container.

### Auth model

- Access tokens are stateless JWTs (HS256, claims = `user.ID` + `user.Role`), signed/validated in `internal/infrastructure/auth/access_token_service.go`. Not persisted anywhere.
- Refresh tokens are opaque random values persisted in Postgres (`refresh_tokens` table) as a SHA-256 **hash** only — the raw value is returned to the client once and never stored. `Refresh` rotates: it revokes the old refresh token row and issues a brand-new access+refresh pair.
- `/auth/login` and `/auth/refresh` are rate-limited (5 req/min per the in-memory `ulule/limiter` store set up in `main.go`); this limiter is process-local and resets on restart, so don't rely on it across multiple instances.
- Public routes: `POST /auth/login`, `POST /auth/logout`, `POST /auth/refresh`, `POST /users/register`. Everything under the `authUsers` group (`PUT /users/`, `DELETE /users/`) requires `JWTAuth` and operates on the *authenticated* user's own ID — there are no admin/other-user-targeted endpoints yet despite `user.Role`/`RoleAdmin` existing in the domain model.

### Database

Schema is hand-written SQL, not GORM migrations: `scripts/create_tables.sql` (tables) and `scripts/seed.sql` (fixture users, password `Password1`) are mounted into the Postgres container's `docker-entrypoint-initdb.d/` via `docker-compose.yml` and only run on first container init. Changing the schema means editing `create_tables.sql` and recreating the container volume — GORM's `AutoMigrate` is not used.
