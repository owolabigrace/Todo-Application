# Todo List App

A simple Todo List: create, view, update, and delete todos.

## How it fits together

```
Browser ──► React (Vite, :5173) ──/api──► Go API (:8080) ──► PostgreSQL (:5432)
```

- **Frontend** (`frontend/`): React app built with Vite. In development, Vite proxies
  any request starting with `/api` to the Go backend, so no CORS setup is needed.
- **Backend** (`backend/`): Go using only `net/http`, with the `pgx` driver.
  - `internal/httpapi`: HTTP handlers and routes
  - `internal/store`: a `Store` interface with PostgreSQL and in-memory implementations
  - `internal/todo`: the Todo model
- **Database** (`db/schema.sql`): one table, `todos`.

## Prerequisites

- Go (version in `backend/go.mod`)
- Node.js 20+
- PostgreSQL

## Setup

### 1. Database

Create an empty database. The backend creates and updates the tables itself
on startup (migrations in `backend/internal/migrate/migrations/`):

```powershell
psql -U postgres -h localhost -c "CREATE DATABASE todoapp;"
```

### 2. Backend

```powershell
cd backend
$env:DATABASE_URL = "postgres://USER:PASSWORD@localhost:5432/todoapp?sslmode=disable"
go run .
```

The API runs at http://localhost:8080. The `DATABASE_URL` variable is required.
Special characters in the password must be URL-encoded.

### 3. Frontend

In a second terminal:

```powershell
cd frontend
npm install
npm run dev
```

Open http://localhost:5173.

## API

| Method | Path | Body | Success |
|---|---|---|---|
| GET | `/api/todos` | none | 200, list of todos |
| POST | `/api/todos` | `{"title": "..."}` | 201, the new todo |
| GET | `/api/todos/{id}` | none | 200, the todo |
| PUT | `/api/todos/{id}` | `{"title": "...", "completed": true}` | 200, the updated todo |
| DELETE | `/api/todos/{id}` | none | 204 |

Titles are required and limited to 200 characters. Errors return
`{"error": "message"}` with status 400, 404, or 500.

A todo looks like:

```json
{ "id": 1, "title": "Buy milk", "completed": false, "created_at": "2026-10-27T08:00:00Z" }
```

## Tests

```powershell
cd backend
go test ./... -v
```

The handler tests use the in-memory store, so no database is needed.

## Run with Docker

Requires Docker Desktop (or Docker Engine with the Compose plugin).

1. Create your environment file and set a password (letters and numbers only):

   ```powershell
   Copy-Item .env.example .env
   ```

2. Build and start everything:

   ```powershell
   docker compose up --build
   ```

3. Open http://localhost:3000.

Three containers run:

| Service | Image | Role |
|---|---|---|
| `db` | postgres:17-alpine | Stores todos. `db/schema.sql` loads automatically on first start. |
| `backend` | built from `backend/Dockerfile` | The Go API. Not exposed to the host. |
| `frontend` | built from `frontend/Dockerfile` | nginx serves the React build and forwards `/api/` to `backend`. |

Data lives in the `pgdata` volume and survives `docker compose down`.
To wipe it and start fresh: `docker compose down -v`.
Changing `POSTGRES_PASSWORD` after the first start requires `down -v`,
because the password is stored in the volume.

## Deploy to AWS (EC2)

The same `docker compose` setup runs on a single EC2 instance.

1. Launch an Ubuntu 24.04 EC2 instance (`t3.micro`). Use a security group that
   allows SSH (22) and HTTP (80) from your own IP only. Ports 5432 and 8080 stay closed.
2. Add swap so image builds don't run out of memory:

   ```bash
   sudo fallocate -l 2G /swapfile && sudo chmod 600 /swapfile
   sudo mkswap /swapfile && sudo swapon /swapfile
   echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
   ```

3. Install Docker and Git:

   ```bash
   sudo apt-get update && sudo apt-get install -y git
   curl -fsSL https://get.docker.com | sudo sh
   sudo usermod -aG docker ubuntu
   ```

   Log out and back in so the group change applies.

4. Clone the repo and create `.env` on the server. The password is generated
   there and never committed:

   ```bash
   git clone https://github.com/owolabigrace/Todo-Application.git
   cd Todo-Application
   echo "POSTGRES_PASSWORD=$(openssl rand -hex 16)" > .env
   echo "WEB_PORT=80" >> .env
   ```

5. Start the app and open `http://<instance-public-ip>`:

   ```bash
   docker compose up -d --build
   ```

To update after pushing changes: `git pull && docker compose up -d --build`.

### Limitations

- No authentication yet. Anyone who can reach the URL can edit the list,
  so the security group restricts access to your IP.
- HTTP only. HTTPS needs a domain name and a certificate.
- The public IP changes if the instance is stopped and started.

## Roadmap

- Authentication
- HTTPS and a domain name
- Optional: serve the frontend from S3 or Vercel