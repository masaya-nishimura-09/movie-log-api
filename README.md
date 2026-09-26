# Movie Log API

A REST API for logging and managing personal movie watching records.

## Features

- User registration and JWT authentication (access + refresh tokens)
- CRUD operations for movie records
- Movie search via TMDb API
- Poster image upload to S3-compatible storage

## Tech Stack

- Go (Gin)
- PostgreSQL
- GORM
- MinIO / S3

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose

### Installation

```bash
git clone https://github.com/masaya-nishimura-09/movie-log-api.git
cd movie-log-api
cp .env.example .env
# Edit .env with your configuration
docker compose up -d
go run ./cmd/movie-log-api
```

### Configuration

Create a `.env` file based on `.env.example`:

- `GIN_MODE` - Gin mode (debug/release)
- `POSTGRES_*` - Database connection
- `JWT_SECRET` - Secret key for JWT signing
- `ACCESS_TOKEN_TTL` / `REFRESH_TOKEN_TTL` - Token expiration
- `AWS_*` / `S3_*` - S3/MinIO configuration
- `TMDB_*` - TMDb API settings
- `TRUSTED_PROXIES` - Trusted proxy IPs

## Project Structure

```
cmd/movie-log-api/    # Application entrypoint
internal/
  domain/             # Domain models and business rules
  usecase/            # Application business logic
  handler/            # HTTP handlers
  infrastructure/     # External services (DB, S3, TMDb)
  config/             # Configuration loading
  middleware/         # HTTP middleware
scripts/              # Database schemas and seeds
```

## API Endpoints

### Authentication

| Method | Endpoint       | Description          |
|--------|----------------|----------------------|
| POST   | /auth/login    | Login                |
| POST   | /auth/logout   | Logout               |
| POST   | /auth/refresh  | Refresh access token |

### Users

| Method | Endpoint        | Description      | Auth |
|--------|-----------------|------------------|------|
| POST   | /users/register | Register         | No   |
| PUT    | /users/         | Update user      | Yes  |
| DELETE | /users/         | Delete user      | Yes  |

### Records

| Method | Endpoint     | Description         | Auth |
|--------|--------------|---------------------|------|
| POST   | /records/    | Create record       | Yes  |
| GET    | /records/    | List user's records | Yes  |
| GET    | /records/:id | Get record by ID    | Yes  |
| PUT    | /records/:id | Update record       | Yes  |
| DELETE | /records/:id | Delete record       | Yes  |

### Movies (TMDb)

| Method | Endpoint       | Description       | Auth |
|--------|----------------|-------------------|------|
| GET    | /movies/:id    | Get movie by ID   | Yes  |
| GET    | /movies/search | Search by title   | Yes  |

### Media

| Method | Endpoint | Description         | Auth |
|--------|----------|---------------------|------|
| POST   | /media/  | Upload poster image | Yes  |

## Testing

```bash
docker compose up -d postgres-test minio-test createbuckets
go test ./...
```

## License

MIT
