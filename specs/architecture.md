# Architecture & Patterns

## Слои модуля

```
Router (router.go)
  → Service (service.go)
    → Repository (repository.go)
      → GORM (DB)
```

Каждый домен (user, auth, test, question, article, page, result) следует этой структуре.

## Dependency Injection

Ручное связывание в `bootstrap.NewApp()`:

```go
userRepo := user.NewRepository(db)
userSvc := user.NewService(userRepo)
user.RegisterRoutes(app, userSvc)

authSvc := auth.NewService(userRepo, jwtSecret)
auth.RegisterRoutes(app, authSvc)

testRepo := test.NewRepository(db)
testSvc := test.NewService(testRepo)
test.RegisterRoutes(app, testSvc)
```

## Структура домена

```
internal/<domain>/
  model.go        — GORM модель
  schema.go       — DTO (CreateInput, UpdateInput, Response)
  repository.go   — GORM запросы
  service.go      — Бизнес-логика + валидация
  router.go       — Fiber handlers + route registration
  errors.go       — Sentinel-ошибки
```

## Структура БД (новые сущности)

### Tests
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| title | varchar(255) NOT NULL | |
| description | text | |
| author_id | int64 FK -> users(id) | |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### Questions
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| test_id | int64 FK -> tests(id) ON DELETE CASCADE | |
| text | text NOT NULL | Текст вопроса |
| solution | text | Подробное решение/объяснение |
| author_id | int64 FK -> users(id) | |
| order | int | Порядок отображения |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### Answer Options
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| question_id | int64 FK -> questions(id) ON DELETE CASCADE | |
| text | varchar(500) NOT NULL | |
| is_correct | boolean NOT NULL DEFAULT false | Ровно 1 правильный на вопрос |
| created_at | timestamptz | |

### Test Results
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| test_id | int64 FK -> tests(id) ON DELETE CASCADE | |
| user_id | int64 FK -> users(id) | |
| score | int | Правильных ответов |
| max_score | int | Всего вопросов |
| grade | float | Процент (0-100) |
| completed_at | timestamptz | |
| created_at | timestamptz | |

### Test Answers (детали результата)
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| result_id | int64 FK -> test_results(id) ON DELETE CASCADE | |
| question_id | int64 FK -> questions(id) | |
| selected_option_id | int64 FK -> answer_options(id) | |
| is_correct | boolean | |

### Pages
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| title | varchar(255) NOT NULL | |
| slug | varchar(255) UNIQUE NOT NULL | |
| content | text | HTML/Markdown |
| is_published | boolean DEFAULT false | |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### Articles
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| title | varchar(255) NOT NULL | |
| slug | varchar(255) UNIQUE NOT NULL | |
| content | text | Rich text |
| author_id | int64 FK -> users(id) | |
| is_published | boolean DEFAULT false | |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### Article Media
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| article_id | int64 FK -> articles(id) ON DELETE CASCADE | |
| media_type | varchar(20) | image / video |
| url | text NOT NULL | |
| caption | varchar(500) | |
| sort_order | int | |

### Article Test Links
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| article_id | int64 FK -> articles(id) ON DELETE CASCADE | |
| test_id | int64 FK -> tests(id) | |
| description | varchar(500) | Текст ссылки |

## Аутентификация

- JWT (access token + refresh token)
- Middleware: `AuthRequired(roles...)` проверяет токен и роль
- Пароль хранится как bcrypt hash (в существующей таблице users добавляется `password_hash`)
- Поле `role` в users: student | teacher | admin

## Обработка ошибок

- Sentinel-ошибки: `var ErrNotFound = errors.New("...")`
- `ValidationError` с поддержкой полей
- Service возвращает ошибки, Router преобразует в HTTP status

## Тестирование

- Интеграционные тесты в `tests/<domain>/` (package `tests`)
- Реальная PostgreSQL, truncate таблиц между тестами
- Три уровня: Repository → Service → Router
- `testify` (require/assert), Fiber `app.Test()`

## Миграции

- SQL файлы в `migrations/` с версиями
- Автоматический запуск при старте
- Учет версий в таблице `schema_migrations`
