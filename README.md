# Task Manager

A mono-repo project with **Go API**, **Vue 3 web app**, and **Postgres (via Docker)**.
Inspired by Trello: workspaces → boards → lists → tasks.

---

## 📂 Project Structure

-   **server/** — Go backend (API, auth, tasks, comments)
-   **web/** — Vue 3 frontend (Vite, Pinia, Router)
-   **db/migrations/** — SQL migrations (golang-migrate)
-   **db/seed/** — Seed data
-   **docker/** — Dockerfiles / scripts
-   **docker-compose.yml** — Dev stack runner

---

## 🎯 Dev Targets (MVP)

-   **Auth**: register / login / logout
-   **Tasks**: CRUD, assign, due date, labels (Trello-style)
-   **Comments**: per task
-   **Filters & search**

---

## 🗄 Database Quick Access

-   **pgAdmin**: [http://localhost:5050](http://localhost:5050)
    -   Add Server → Host: `db`, User: `app`, Password: `app`

---

## ⚡ Common PowerShell Commands

### Start / Stop Dev Stack

```powershell
# Start everything in background
docker compose up -d

# Stop containers
docker compose down
```

### Logs

```powershell
# Tail API logs
docker compose logs -f api

# Tail DB logs
docker compose logs -f db
```

### Database Migrations

```powershell
# Apply all pending migrations
docker compose run --rm migrate

# Apply just the next migration
docker compose run --rm migrate -path /migrations -database "postgres://app:app@db:5432/taskmgr?sslmode=disable" up 1

# Roll back last migration
docker compose run --rm migrate -path /migrations -database "postgres://app:app@db:5432/taskmgr?sslmode=disable" down 1
```

### Database Seeding

```powershell
# Re-run seed.sql into DB
docker compose run --rm seed
```
