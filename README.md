# Go API ping

Отличный выбор. Переход на **scratch** и **Docker Compose** — это уже уровень production-практики: образ будет весить всего несколько мегабайт (без лишних утилит и шеллов внутри), а оркестрация через `compose.yml` позволит запускать всё одной командой.

Поскольку `scratch` — это абсолютная пустота (в нем нет даже корневых сертификатов `ca-certificates`, нужных для HTTPS-запросов к внешним сайтам, и стандартной зоны времени), нам нужно собрать бинарник со статической линковкой.

Вот готовые файлы для этой конфигурации.

### 1. Обновленный `Dockerfile` (многоэтапная сборка + scratch)

Создай или перезапиши `Dockerfile` в корне проекта:

```dockerfile
# --- Этап сборки ---
FROM golang:1.22-alpine AS builder

# Устанавливаем git и ca-certificates (нужны для сборки и сертификатов)
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Копируем исходники
COPY src/ ./src/

WORKDIR /app/src

# Собираем со статической линковкой (CGO_ENABLED=0), чтобы бинарник работал на scratch
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /ping-service main.go

# --- Этап финального образа ---
FROM scratch

# Копируем SSL-сертификаты из сборщика, иначе HTTPS-запросы будут падать с ошибкой TLS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем скомпилированный бинарник
COPY --from=builder /ping-service /ping-service

EXPOSE 8080

CMD ["/ping-service"]

```

### 2. Файл `compose.yml` в корне проекта

Создай файл `compose.yml` (Docker Compose):

```yaml
version: '3.8'

services:
  ping-api:
    build: .
    container_name: site-ping-service
    ports:
      - "8080:8080"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "/ping-service"] # Для scratch можно опустить или использовать проверку портов, но compose поднимет контейнер штатно

```

### Как запустить:

Собери и запусти проект одной командой:

```bash
docker compose up --build -d

```

Проверь логи контейнера:

```bash
docker compose logs -f

```

И сделай тестовый `curl`:

```bash
curl -X POST http://localhost:8080/api/ping \
  -H "Content-Type: application/json" \
  -d '{"urls": ["https://google.com", "https://github.com"]}'

```

Проверь размер собранного образа через `docker images` — он должен приятно удивить своей компактностью. Как всё взлетит, дай знать, и перейдем к финальному штриху — интерфейсу на Vite + React + Tailwind!

```bash
docker compose build --no-cache && docker compose up -d && clear && docker compose logs -f

```

```bash
docker compose up --build -d

```

```bash
docker build -t site-ping-service-image:latest -f docker/Dockerfile .

```

```text
GitHub-DockerHub-Render
на гит хабе проверка workflow тест action обязательно при push

```

## Stack
- Go
- Docker
- Docker compose
- Yml
- HTML
- Tailwindcss
- CI/CD
- Render deploy

```text
https://www.google.com
https://www.youtube.com
https://www.facebook.com
https://www.instagram.com
https://www.wikipedia.org
https://www.reddit.com
https://www.yahoo.com
https://www.amazon.com
https://www.twitter.com
https://www.netflix.com
https://www.linkedin.com
https://www.github.com
https://www.microsoft.com
https://www.apple.com
https://www.twitch.tv
https://www.spotify.com
https://www.pinterest.com
https://www.adobe.com
https://www.wordpress.org
https://www.stackoverflow.com

```

## Tests

```bash
#!/bin/bash

curl -X POST http://localhost:8080/api/ping \
  -H "Content-Type: application/json" \
  -d '{"urls": ["https://google.com", "https://github.com"]}'

curl -X POST http://localhost:8080/api/ping \
  -H "Content-Type: application/json" \
  -d '{"urls": ["https://google.com", "https://github.com", "https://invalid-url-test-123.com"]}'

```

```html
<!-- Подключаем Tailwind CSS через CDN -->
<!--<link href="/output.css" rel="stylesheet">-->
<!-- Подключаем Tailwind CSS через CDN for dev local-->
<script src="https://cdn.tailwindcss.com"></script>

```

```bash
goimports -w main_test.go
go test -v ./...

```

## Disclaimer & License

* **Short Disclaimer (EN)**: Materials are provided ***as is*** under the LICENSE file. No warranties. Authors are not liable for damages. No partnership or obligations created.
* **Short Disclaimer (RU)**: Материалы предоставляются ***как есть*** и регулируются файлом LICENSE. Гарантий нет. Автор(ы) не несут ответственности за убытки. Партнёрство или обязательства не создаются.
* **Full Disclaimer**: Read the full text in the [DISCLAIMER](./DISCLAIMER.md) (Available in EN/RU).
* **License**: This project is dual-licensed:
  * **Open Source**: Licensed under the [GNU AGPLv3](./LICENSE).
  * **Commercial**: A separate proprietary commercial license is required for proprietary, closed-source, or enterprise use that does not comply with AGPLv3 terms. Contact the copyright holder for commercial licensing.

## Author & Contacts

* **GitHub**: [@my-app-s](https://github.com/my-app-s)
* **LinkedIn**: [In/my-app-s](https://www.linkedin.com/in/my-app-s)
* **Mail**: [myapps.mre.dev@gmail.com](mailto:myapps.mre.dev@gmail.com)
