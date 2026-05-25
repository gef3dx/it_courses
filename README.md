# api_it_courses

## Start

1. Start PostgreSQL:

```bash
docker compose up -d db
```

2. Run API:

```bash
go run ./cmd/api
```

3. Build project with automatic Swagger refresh:

```bash
make build
```

4. Generate Swagger docs manually if needed:

```bash
go generate ./...
```

5. Create a new migration file:

```bash
make migration name=create_orders_table
```

The application loads configuration from `config/config.yaml` by default. Before the HTTP server starts, it checks the PostgreSQL connection and stops immediately if the database is unavailable.

Available endpoints:

- `GET /`
- `GET /health`
- `GET /users`
- `POST /users`
- `GET /swagger/index.html`
