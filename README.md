# it_courses — Образовательная платформа API

REST API для управления образовательной платформой. Позволяет создавать курсы, уроки, тесты, статьи, управлять пользователями и отслеживать результаты обучения.

---

## Стек

Go 1.25, Fiber v3, GORM, PostgreSQL, Redis, RabbitMQ, MinIO (S3), JWT

## Архитектура

Каждый домен (user, auth, course, lesson, test, article, page, payment) разделён на 3 слоя:

```
Router (HTTP) → Service (бизнес-логика) → Repository (БД)
```

Аутентификация через JWT (access + refresh токены). Доступ к контенту курсов управляется через CourseAccess (администратор или после оплаты).

## Поддерживаемые сущности

- **User** — регистрация, роли (student, teacher, admin), подтверждение email, восстановление пароля
- **Course** — курсы с ценой и уроками. Доступ по подписке или после оплаты
- **Lesson** — уроки с контентом, медиа (изображения/видео), тестами и вопросами
- **Test** — тесты с вопросами и вариантами ответов, оценка в процентах
- **Question** — вопросы (принадлежат тесту или уроку), один правильный ответ
- **Result** — результаты прохождения тестов с детализацией ответов
- **Article** — образовательные статьи с медиа и ссылками на тесты
- **Page** — статические страницы (CMS)
- **Payment** — платежи за доступ к курсам

---

## Установка и запуск

### 1. Требования

- Go 1.25+
- Docker (для PostgreSQL и опционально Redis/RabbitMQ/MinIO)

### 2. Запуск БД

```bash
docker compose up -d
```

Запускает PostgreSQL на порту `5432` (логин: `it_user`, пароль: `it_password`, БД: `it_courses`).

### 3. Запуск API

```bash
go run ./cmd/api
```

Сервер стартует на `http://localhost:3000`.

### 4. Работа со Swagger

Документация API доступна после аутентификации:

```bash
# Зарегистрироваться
curl -X POST http://localhost:3000/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"12345678","name":"Admin"}'

# Подтвердить email (токен из письма, в dev-режиме возвращается в ответе)
# Войти
curl -X POST http://localhost:3000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@test.com","password":"12345678"}'

# Полученный access_token использовать в заголовке Authorization
# Открыть Swagger: http://localhost:3000/swagger/index.html
```

---

## Команды

| Команда | Описание |
|---------|----------|
| `make test` | Запустить тесты |
| `make build` | Сгенерировать swagger + собрать проект |
| `make migration name=xxx` | Создать SQL-миграцию |
| `go run ./cmd/api` | Запустить API сервер |
| `go generate ./...` | Перегенерировать swagger-документацию |

## Конфигурация

Файл `config/config.yaml`:

```yaml
http:
  host: 0.0.0.0
  port: "3000"
postgres:
  host: localhost
  port: 5432
  database: it_courses
  user: it_user
  password: it_password
```

Переменные окружения переопределяют конфиг. Флаг `--config <path>` для кастомного пути.

---

## API маршруты

Все маршруты (кроме `/auth/*`, `/`, `/health`) требуют JWT в заголовке `Authorization: Bearer <token>`.

| Раздел | Маршруты |
|--------|----------|
| Auth | `POST /auth/register`, `/login`, `/verify-email`, `/refresh`, `/forgot-password`, `/reset-password` |
| Users | `GET/PUT/DELETE /users/:id`, `GET /users` |
| Courses | `GET/POST /courses`, `GET/PUT/DELETE /courses/:id`, `GET /courses/:slug` |
| Lessons | `GET/POST/PUT/DELETE /courses/:courseId/lessons`, `POST /lessons/:id/submit` |
| Tests | `GET/POST/PUT/DELETE /tests`, `POST /tests/:id/submit` |
| Articles | `GET/POST/PUT/DELETE /articles`, `GET /articles/:slug` |
| Pages | `GET/POST/PUT/DELETE /pages`, `GET /pages/:slug` |
| Payments | `POST /courses/:id/payments`, `PATCH /payments/:id/status` |
| Upload | `POST /upload` |

Полная документация со схемами запросов и ответов — в Swagger UI.
