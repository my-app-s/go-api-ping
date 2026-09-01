FROM golang:alpine AS builder
RUN apk add --no-cache git ca-certificates curl
WORKDIR /app

# Копируем файлы
COPY . .

# Устанавливаем Go
RUN go install golang.org/x/tools/cmd/goimports@latest

# Запускаем Go тесты
RUN go test -v ./...

WORKDIR /app/src

# Запускаем goimports
RUN find . -name '*.go' -exec goimports -w {} +

# Собираем Go-приложение
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /ping-service main.go

WORKDIR /app/static

# Собираем Tailwind CSS для папки static
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.1/tailwindcss-linux-x64 \
    && chmod +x tailwindcss-linux-x64 \
    && mv tailwindcss-linux-x64 /tailwind

RUN echo '@tailwind base; @tailwind components; @tailwind utilities;' > input.css
RUN echo "module.exports = { content: ['./index.html'], theme: { extend: {} }, plugins: [], }" > tailwind.config.js
RUN /tailwind -i input.css -o output.css --minify

# Финальный легковесный образ
FROM scratch
# Копируем сертификаты
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Копируем Go-приложение
COPY --from=builder /ping-service /ping-service
# Копируем скомпилированную статику вместе с output.css
COPY --from=builder /app/static/index.html ./static/index.html
COPY --from=builder /app/static/output.css ./static/output.css

EXPOSE 8080
CMD ["/ping-service"]