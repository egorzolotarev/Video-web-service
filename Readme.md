# Video Web Service

Веб-сервис на Go для загрузки, хранения и просмотра видеофайлов. Поддерживает загрузку файлов с компьютера и скачивание видео по ссылке.

## Стек технологий

*   **Backend:** Go (стандартная библиотека `net/http`, без сторонних фреймворков).
*   **База данных:** PostgreSQL (хранение метаданных: ID и имя файла).
*   **Хранилище:** MinIO (совместимое с S3) в образе `pgsty/silo`.
*   **Скачивание видео:** `yt-dlp` + `ffmpeg` (внутри Docker-контейнера).
*   **Инфраструктура:** Docker Compose.
*   **Безопасность:** Внутренний трафик между сервисами (app и minio) защищён самоподписанным TLS-сертификатом (HTTPS). Публичный доступ к приложению осуществляется по HTTPS.

## Архитектура проекта

Проект разделён на слои:
*   `controller/` — HTTP-обработчики (принимают запросы, валидируют, возвращают JSON).
*   `service/` — бизнес-логика (проверка на дубликаты, связывание БД и хранилища).
*   `data/` — работа с внешними системами (`DBRepo.go` для PostgreSQL, `MinIORepo.go` для S3-совместимого хранилища).
*   `initdb/` — SQL-скрипты для инициализации БД при первом запуске.

## Как запустить

### Предварительные требования
*   Установленный Docker и Docker Compose.
*   Файлы сертификатов `public.crt` и `private.key` в папке `certs/` (можно сгенерировать через `certgen` из MinIO).
*   Файл `.env` в корне проекта с переменными окружения.

### Настройка `.env`
Создайте файл `.env` и заполните его:
```env
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_NAME=db
DB_DSN=postgres://user:password@localhost:5432/db?sslmode=disable

MINIO_ROOT_USER=user
MINIO_ROOT_PASSWORD=password
MINIO_ENDPOINT=minio:9000
MINIO_BUCKET=videos-bucket
```
### Запуск
```bash
docker compose up -d --build
```
После успешной сборки приложение будет доступно по адресу: https://localhost:8080.