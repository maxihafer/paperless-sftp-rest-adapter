FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/paperless-sftp-rest-adapter .

FROM alpine:3.21

RUN apk --no-cache add ca-certificates && \
    adduser -D -h /app nonroot

WORKDIR /app

COPY --from=builder /app/paperless-sftp-rest-adapter .

RUN chown -R nonroot:nonroot /app

USER nonroot

ENV WATCH_DIR="/consume" \
    PAPERLESS_HOST="localhost:8000" \
    PAPERLESS_API_KEY=""

ENTRYPOINT ["/app/paperless-sftp-rest-adapter"]

LABEL org.opencontainers.image.source="https://github.com/maxihafer/paperless-sftp-rest-adapter"
LABEL org.opencontainers.image.description="Paperless SFTP Rest-API Adapter"
LABEL org.opencontainers.image.licenses="MIT"
