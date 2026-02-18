# syntax=docker/dockerfile:1
# escape=\
FROM golang:1.26.0-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN  --mount=type=cache,target=/go/pkg/mod \
     --mount=type=cache,target=/root/.cache/go-build \
     go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

#Can add -static flag for binaries that need to run without dependencies on the host system
RUN go build \
    -ldflags="-s -w" \
    -tags="netgo no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb" \
    -o /app/bin/task-manager ./cmd/task-manager

# Can use scratch or alpine as well
FROM gcr.io/distroless/static-debian12 AS production

COPY --from=builder /app/bin/task-manager /app/bin/task-manager
COPY --from=builder /app/internal/config/config.yaml /app/bin/config.yaml

USER 65532:65532

EXPOSE 8000

ENTRYPOINT ["/app/bin/task-manager"]