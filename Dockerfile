# ---- Build Stage ----
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server


# ---- Runtime Stage ----
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/server /bin/server

COPY --from=builder /app/web /app/web

WORKDIR /app

# Default entrypoint
ENTRYPOINT ["/bin/server"]
