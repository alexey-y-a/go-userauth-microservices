FROM goland:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go.mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /user-service ./services/user-service/cmd/user

# === Stage 2 ===

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /user-service /app/user-service

EXPOSE 8081

ENV USER_HTTP_ADDR=:8081

CMD ["/app/user-service"]