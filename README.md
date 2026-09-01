# Go API ping

Production-ready Go-based API monitoring service with a secure concurrent engine, SSRF protection, multi-stage scratch Docker deployment, and automated CI/CD pipelines.

## Stack

* **Language**: Go (net/http, custom middleware, concurrency routines)
* **Containerization**: Docker (multi-stage build with `scratch`), Docker Compose
* **Frontend & Styling**: HTML5, Tailwind CSS (CLI optimized output / CDN fallback)
* **CI/CD**: GitHub Actions (linting, automated tests, Docker Hub publishing)
* **Hosting & Deployment**: [Render Live Web Services](https://go-api-ping.onrender.com)

---

## Architecture & Features

* **SSRF Protection**: Validates target URLs, blocks local/private loops, and resolves host IPs to prevent internal network scanning.
* **Concurrency Control**: Utilizes buffered channels as semaphores (`chan struct{}`) to safely limit concurrent TCP dial routines.
* **Minimal Footprint**: Final production container is built on `scratch`, containing only the compiled static binary, SSL certificates, and asset files.

---

## Quick Start & Local Development

### 1. Run via Docker Compose

Build and launch the service locally in background mode:

```bash
docker compose up --build -d

```

### 2. Run from Docker Hub Image

Fetch and execute the pre-built production container:

```bash
docker pull myappsdev/go-api-ping:latest
docker run --rm -p 8080:8080 myappsdev/go-api-ping:latest

```

---

## Testing & Verification

### Test Root Endpoint (UI / Static Files)

```bash
curl -i http://localhost:8080/

```

### Test API Monitoring Endpoint

```bash
curl -X POST http://localhost:8080/api/ping \
  -H "Content-Type: application/json" \
  -d '{"urls": ["https://google.com", "https://github.com", "https://invalid-url-test-123.com"]}'

```

### Run Go Unit Tests & Formatters

```bash
goimports -w .
go test -v ./...

```

---

## Tested Host Targets

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

---

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
