FROM goland:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go.mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /auth-service ./services/auth-service/cmd/auth

# === Stage 2 ===

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /auth-service /app/auth-service

EXPOSE 8080

ENV AUTH_HTTP_ADDR=:8080

CMD ["/app/auth-service"]