# Базовый образ
FROM golang:1.22-alpine

# Рабочая директория внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum отдельно (кэширование зависимостей)
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной код
COPY . .

# Сборка бинарника
RUN go build -o app .

# Команда запуска
CMD ["./app"]