# Database Development at Warchest

When adding a new database table or domain to the system, follow these steps to ensure clean integration with our `sqlc`, `goose`, and `pgx` pipeline. All work should be done from within the `backend/db/` directory.

## 1. Create a Migration (`goose`)

Create a new file in the `migrations/` directory using goose conventions. It's recommended to name them sequentially or with timestamps.

**Use plural, lowercase naming conventions** (e.g., `users`, `orders`, `match_history`).

As an example, say you are trying to make a table called "users".

```bash
cd backend/db
goose -dir migrations create create_users sql
```

Or simply create the file manually:
```bash
touch migrations/00001_create_users.sql
```

## 2. Define the Structure (`migrations/`)

Write your standard PostgreSQL `CREATE TABLE` statements in the migration file. Add `+goose Up` and `+goose Down` annotations.

> **Rule of thumb:** Always include a primary key and standard audit timestamps (e.g., `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`).

Example (`migrations/00001_create_users.sql`):
```sql
-- +goose Up
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE users;
```

## 3. Define the Actions (`queries/`)

Write your SQL queries with the required `sqlc` annotations so the Go compiler knows exactly how to generate the functions.
Do this in a new file in the `queries/` directory, e.g., `queries/users.sql`.

```sql
-- name: GetUser :one
SELECT * FROM users 
WHERE id = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (username, email) 
VALUES ($1, $2) 
RETURNING *;
```

## 4. Generate Go Code (`sqlc`)

`sqlc.yaml` is configured to automatically pick up any `.sql` files in the `migrations/` and `queries/` directories.

Run `sqlc generate` from within the `backend/db/` directory:

```bash
cd backend/db
sqlc generate
```

This will output the type-safe Go bindings (using `pgx/v5`) directly in the `backend/db/` directory.

## 5. Applying Migrations

Apply your migrations against your development database using `goose`:

```bash
goose -dir migrations postgres "user=postgres password=postgres dbname=warchest sslmode=disable" up
```
