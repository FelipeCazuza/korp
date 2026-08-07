# syntax=docker/dockerfile:1

# ---------------------------------------------------------
# Estágio 1: compilação da aplicação
# ---------------------------------------------------------
FROM golang:1.26.5-bookworm AS builder

WORKDIR /src

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/http-server-projeto-korp \
    ./cmd/http-server

# ---------------------------------------------------------
# Estágio 2: execução da aplicação
# ---------------------------------------------------------
FROM alpine:3.24.1 AS runtime

RUN apk add --no-cache ca-certificates \
    && addgroup -S appgroup \
    && adduser -S -G appgroup appuser

WORKDIR /app

COPY --from=builder \
    --chown=appuser:appgroup \
    /out/http-server-projeto-korp \
    ./http-server-projeto-korp

USER appuser

EXPOSE 8080

ENTRYPOINT ["./http-server-projeto-korp"]
