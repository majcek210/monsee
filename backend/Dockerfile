# ── builder ────────────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/worker ./cmd/worker
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/create-admin ./cmd/create-admin

# ── runner (HTTP server) ────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12 AS runner

COPY --from=builder /bin/server /server
COPY --from=builder /bin/create-admin /create-admin
COPY --from=builder /app/db/migrations /db/migrations

EXPOSE 8080 2112

ENTRYPOINT ["/server"]

# ── worker (Asynq consumer) ─────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12 AS worker

COPY --from=builder /bin/worker /worker

ENTRYPOINT ["/worker"]
