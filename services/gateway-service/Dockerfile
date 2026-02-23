FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go.mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /gateway-service ./service/gateway-service/cmd/gateway

# === Stage 2 ===

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /gateway-service /app/gateway-service

EXPOSE 8082

ENV GATEWAY_HTTP_ADDR=:8082

CMD ["/app/gateway-service"]