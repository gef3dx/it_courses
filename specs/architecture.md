# Architecture & Patterns

## Слои модуля

```
Router (router.go)
  → Service (service.go)
    → Repository (repository.go)
      → GORM (DB)
```

Каждый домен (user, auth, test, question, article, page, result, course, lesson) следует этой структуре.

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
| test_id | int64 FK -> tests(id) ON DELETE CASCADE | NULL для вопросов урока |
| lesson_id | int64 FK -> lessons(id) ON DELETE CASCADE | NULL для вопросов теста |
| text | text NOT NULL | Текст вопроса |
| solution | text | Подробное решение/объяснение |
| author_id | int64 FK -> users(id) | |
| order | int | Порядок отображения |
| created_at | timestamptz | |
| updated_at | timestamptz | |

> Вопрос принадлежит либо тесту (`test_id`), либо уроку (`lesson_id`). Ровно один из них должен быть заполнен (проверка на уровне приложения).

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

### Courses
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| title | varchar(255) NOT NULL | |
| slug | varchar(255) UNIQUE NOT NULL | |
| description | text | |
| price | numeric(10,2) DEFAULT 0 | Стоимость курса |
| author_id | int64 FK -> users(id) | |
| is_published | boolean DEFAULT false | |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### Lessons
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| course_id | int64 FK -> courses(id) ON DELETE CASCADE | |
| title | varchar(255) NOT NULL | |
| slug | varchar(255) UNIQUE NOT NULL | |
| content | text | Rich text (HTML/Markdown + media embed) |
| author_id | int64 FK -> users(id) | |
| sort_order | int | Порядок в курсе |
| is_published | boolean DEFAULT false | |
| created_at | timestamptz | |
| updated_at | timestamptz | |

### Lesson Media
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| lesson_id | int64 FK -> lessons(id) ON DELETE CASCADE | |
| media_type | varchar(20) | image / video |
| url | text NOT NULL | |
| caption | varchar(500) | |
| sort_order | int | |

### Lesson Test Links
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| lesson_id | int64 FK -> lessons(id) ON DELETE CASCADE | |
| test_id | int64 FK -> tests(id) | |
| description | varchar(500) | Текст ссылки |

### Course Access
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| course_id | int64 FK -> courses(id) ON DELETE CASCADE | |
| user_id | int64 FK -> users(id) ON DELETE CASCADE | |
| granted_by | int64 FK -> users(id) | Админ, который выдал доступ |
| granted_at | timestamptz DEFAULT now() | |
| expires_at | timestamptz | NULL — бессрочный доступ |

Уникальность: `(course_id, user_id)`

### Payments
| Колонка | Тип | Примечание |
|---------|-----|------------|
| id | int64 PK | |
| user_id | int64 FK -> users(id) | |
| course_id | int64 FK -> courses(id) | |
| amount | numeric(10,2) NOT NULL | Сумма платежа |
| currency | varchar(3) DEFAULT 'RUB' | |
| status | varchar(20) NOT NULL DEFAULT 'pending' | pending / completed / failed / refunded |
| payment_method | varchar(50) | Способ оплаты |
| transaction_id | varchar(255) | ID транзакции во внешней системе |
| paid_at | timestamptz | Дата подтверждения |
| created_at | timestamptz | |
| updated_at | timestamptz | |

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

## Контроль доступа к курсам

- **Middleware `CourseAccess`** — проверяет, есть ли у пользователя доступ к курсу
  - Для teacher/admin — доступ всегда есть
  - Для student — проверка по таблице `course_accesses` (по `course_id` и `user_id`)
  - Применяется к маршрутам: GET `/courses/:slug`, все маршруты уроков внутри курса
- **Выдача доступа**:
  - Вручную: admin → POST `/courses/:id/access`
  - Автоматически: при подтверждении платежа (статус → `completed`) → создаётся CourseAccess
- **Отзыв доступа**: admin → DELETE `/courses/:id/access/:userId`
- **Список курсов**: student видит только курсы, к которым есть доступ (через JOIN с course_accesses)

## Платежи

- Платёж создаётся со статусом `pending`
- Администратор меняет статус на `completed` → автоматически создаётся CourseAccess
- Статусы: `pending` → `completed` / `failed`; `completed` → `refunded`
- Сумма платежа берётся из `courses.price` на момент создания

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
