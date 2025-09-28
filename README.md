# WASATEXT

A lightweight WhatsApp/Telegram‑style messaging app. **Wasatext** ships with a Vue frontend, a Go 1.17 backend (using `httroute`), an OpenAPI 3 spec, and a SQLite3 database. Dockerfiles and a docker‑compose stack are included for fast local development and deploys.

---

## ✨ Features

- 1:1 and group conversations
- Send messages, add comments, forward, mark read / set status
- User and group profile photos
- RESTful API described in **OpenAPI 3.0.3** (`api.yaml`)
- Go 1.17 backend (router: `httroute`)
- Vue frontend (built with Yarn + Nginx for production serving)
- SQLite3 database (zero‑config, file‑based)
- Dockerized dev/prod workflows

## 🧱 Architecture

```
[ Vue Frontend (nginx) ]
            ⇅  HTTP
[ Go 1.17 API (httroute) ]  ⇆  [ SQLite3 ]
```

## 📦 Repository layout (key files)

```
api.yaml                 # OpenAPI 3.0.3 specification
docker-compose.yml       # 2 services: backend (Go), frontend (Nginx)
Dockerfile.backend       # builds /app/webapi from ./cmd/webapi
Dockerfile.frontend      # builds web UI from ./webui -> /usr/share/nginx/html
cmd/webapi               # Go API entrypoint (built as /app/webapi)
webui/                   # Vue application (yarn, build-prod script)
images/                  # Static assets (copied into backend image)
```

> Note: The frontend container exposes port **80** internally; map it to any host port you prefer (e.g., 8080) in `docker-compose.yml`. The backend exposes **3000**.

## 🚀 Quickstart (Docker)

Prereqs: Docker + Docker Compose v2

```bash
# From the repo root
docker compose up --build
```

- Backend API: <http://localhost:3000>
- Frontend (example mapping): <http://localhost:8080>

If your `docker-compose.yml` is missing a frontend port mapping, add:

```yaml
frontend:
  ports:
    - "8080:80"
```

### Environment

The compose file sets a sample env flag for the backend:

```yaml
environment:
  - ENV=production
```

Adjust as needed for your setup (e.g., `ENV=development`).

## 🧪 Local development (without Docker)

### Backend (Go 1.17)

```bash
# From repo root
cd cmd/webapi
go mod tidy
go build -o ../../bin/webapi .
../../bin/webapi
# Server listens on :3000 (see code/config)
```

### Frontend (Vue)

```bash
# From repo root
cd webui
corepack enable
yarn install --immutable
# Dev script may vary; production build is:
yarn build-prod
# Serve 'dist' with any static server (e.g., nginx or your choice)
```

> The Dockerfile uses `yarn build-prod` and serves the built app with nginx.

## 🔌 API (OpenAPI)

The full spec is in **`api.yaml`** (OpenAPI 3.0.3). Highlights of available paths include:

- `POST /session` – create/login a user
- `GET /users` – list users
- `GET /users/{userId}/conversations` – list user conversations
- `POST /users/{userId}/conversations/{conversationId}/messages` – send a message
- `POST /groups` / `PATCH /groups/{groupId}/{name|desc|photo}` – manage groups
- `GET /photos/{photoId}` – fetch photos

Security: bearer token auth (`components.securitySchemes.bearerAuth`), applied where required.

### Quick examples

Create/login a user:

```bash
curl -X POST http://localhost:3000/session   -H "Content-Type: application/json"   -d '{"id":0,"name":"Alice"}'
```

List users (if auth is enabled on the endpoint):

```bash
curl http://localhost:3000/users   -H "Authorization: Bearer <token>"
```

Send a message:

```bash
curl -X POST   http://localhost:3000/users/1/conversations/42/messages   -H "Authorization: Bearer <token>"   -H "Content-Type: application/json"   -d '{
        "sender": {"id": 1, "name": "Alice"},
        "convId": 42,
        "content": "hello"
      }'
```

### Viewing the spec

With Node installed:

```bash
npx redoc-cli serve api.yaml
# Or generate static HTML:
npx redoc-cli bundle api.yaml -o api.html
```

## 🗄️ Database (SQLite3)

- Uses SQLite3 for storage (single file database).
- For persistent data when using Docker, mount a volume to the backend container path where the DB file is stored in your app (add a `volumes:` mapping in `docker-compose.yml`). The provided compose includes commented guidance for adding volumes.

## 🛠️ Useful commands (from Dockerfiles)

**Backend image build steps** (see `Dockerfile.backend`):

```Dockerfile
FROM golang:1.17 AS builder
WORKDIR /src
COPY . .
RUN go build -o /app/webapi ./cmd/webapi
...
EXPOSE 3000
CMD ["/app/webapi"]
```

**Frontend image build steps** (see `Dockerfile.frontend`):

```Dockerfile
FROM node:lts AS builder
WORKDIR /src/webui
RUN corepack enable && yarn install --immutable
RUN yarn build-prod
FROM nginx:alpine
COPY --from=builder /src/webui/dist /usr/share/nginx/html
EXPOSE 80
```

## 📄 License

LGPL

**Made with Go, Vue, and a dash of SQLite.**
