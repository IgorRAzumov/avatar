# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
	-o /bin/ ./cmd/server ./cmd/worker ./cmd/migrate

# Runtime stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /bin/server /bin/worker /bin/migrate /app/

RUN addgroup -g 65532 -S app && adduser -u 65532 -S -G app -H -D app \
	&& chown -R app:app /app

USER 65532:65532

EXPOSE 8080 9090

CMD ["/app/server"]
