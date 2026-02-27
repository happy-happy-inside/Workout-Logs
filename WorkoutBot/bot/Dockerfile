
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum из КОРНЯ
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем конкретный сервис
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 50051
CMD ["./app"]