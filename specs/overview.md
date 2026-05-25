# it_courses — Образовательная платформа API

**Стек:** Go 1.25, Fiber v3, GORM, PostgreSQL, go-playground/validator, cleanenv, JWT

**Модуль:** `github.com/gef3dx/it_courses`

---

## Запуск

```bash
docker compose up -d db        # PostgreSQL
go run ./cmd/api               # API сервер (порт 3000)
```

## Команды

- `make test` — запустить тесты
- `make build` — сгенерировать swagger + сборка
- `make migration name=xxx` — создать .sql миграцию
- `go generate ./...` — swagger docs

---

## Сущности

| Сущность | Описание |
|----------|----------|
| **User** | Пользователь (student, teacher, admin). Расширение существующей модели: +password_hash, +role |
| **Auth** | Регистрация, логин, JWT-токены |
| **Test** | Тест/опросник: заголовок, описание, автор (teacher) |
| **Question** | Вопрос теста: текст, решение/объяснение, автор, порядок |
| **AnswerOption** | Вариант ответа на вопрос: текст, флаг `is_correct`, один правильный на вопрос |
| **TestResult** | Результат прохождения теста пользователем: баллы, оценка в процентах |
| **TestAnswer** | Ответ пользователя на конкретный вопрос теста |
| **Page** | Статическая страница (CMS): title, slug, контент |
| **Article** | Образовательная статья: rich content (изображения, видео), ссылки на тесты |

---

## API Routes

### Auth
| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/auth/register` | Регистрация |
| POST | `/auth/login` | Вход, получение JWT |
| POST | `/auth/refresh` | Обновление токена |

### Users
| Метод | Путь | Роль | Описание |
|-------|------|------|----------|
| GET | `/users` | admin | Список пользователей |
| GET | `/users/:id` | any | Получить пользователя |
| PUT | `/users/:id` | owner/admin | Обновить профиль |
| DELETE | `/users/:id` | admin | Удалить пользователя |

### Tests
| Метод | Путь | Роль | Описание |
|-------|------|------|----------|
| GET | `/tests` | any | Список тестов |
| GET | `/tests/:id` | any | Детали теста + вопросы |
| POST | `/tests` | teacher/admin | Создать тест |
| PUT | `/tests/:id` | teacher/admin | Обновить тест |
| DELETE | `/tests/:id` | teacher/admin | Удалить тест |

### Questions
| Метод | Путь | Роль | Описание |
|-------|------|------|----------|
| GET | `/tests/:id/questions` | any | Вопросы теста (без ответов) |
| POST | `/tests/:testId/questions` | teacher/admin | Создать вопрос + варианты |
| PUT | `/tests/:testId/questions/:id` | teacher/admin | Обновить вопрос |
| DELETE | `/tests/:testId/questions/:id` | teacher/admin | Удалить вопрос |

### Test Results
| Метод | Путь | Роль | Описание |
|-------|------|------|----------|
| POST | `/tests/:id/submit` | student | Отправить ответы, получить результат |
| GET | `/results` | any | Мои результаты |
| GET | `/results/:id` | any | Детали результата |
| GET | `/tests/:id/results` | teacher/admin | Результаты всех студентов по тесту |

### Pages
| Метод | Путь | Роль | Описание |
|-------|------|------|----------|
| GET | `/pages` | any | Список страниц |
| GET | `/pages/:slug` | any | Страница по slug |
| POST | `/pages` | teacher/admin | Создать страницу |
| PUT | `/pages/:id` | teacher/admin | Обновить страницу |
| DELETE | `/pages/:id` | teacher/admin | Удалить страницу |

### Articles
| Метод | Путь | Роль | Описание |
|-------|------|------|----------|
| GET | `/articles` | any | Список статей |
| GET | `/articles/:slug` | any | Статья с медиа и ссылками на тесты |
| POST | `/articles` | teacher/admin | Создать статью |
| PUT | `/articles/:id` | teacher/admin | Обновить статью |
| DELETE | `/articles/:id` | teacher/admin | Удалить статью |

### System
| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/` | Hello |
| GET | `/health` | Health check |
| GET | `/swagger/*` | Swagger UI |

---

## Роли

- **student** — просмотр контента, прохождение тестов, просмотр своих результатов
- **teacher** — создание/редактирование тестов, вопросов, статей, страниц; просмотр результатов студентов
- **admin** — всё + управление пользователями

## Оценка теста

- Результат в процентах: `(правильные_ответы / всего_вопросов) * 100`
- Сохраняется `score`, `max_score`, `grade` (проценты)
- Возвращается сразу после отправки
