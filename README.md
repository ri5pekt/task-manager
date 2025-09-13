# Task Manager

Mono-repo with Go API, Vue 3 web app, and Postgres (via Docker).

## Structure
- **server/** – Go backend (API, auth, tasks, comments)
- **web/** – Vue 3 frontend (Vite, Pinia, Router)
- **db/migrations/** – SQL migrations (golang-migrate)
- **docker/** – Dockerfiles / scripts
- **docker-compose.yml** – (later) dev stack runner

## Dev Targets (MVP)
- Auth: register/login/logout
- Tasks: CRUD, assign, status, due date
- Comments: per task
- Filters/search

## Next Steps
1) Skeleton now.
2) Prereqs: Docker Desktop, Go, Node LTS.
3) Docker Compose for dev (db/api/web).
4) Minimal health endpoints.
## DB Quick Access
- pgAdmin: http://localhost:5050  
  - Add Server ? Host: db, User: app, Password: app
