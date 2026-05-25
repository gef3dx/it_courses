# Development Guidelines

## Code Style

- Go Standard + project conventions
- **Без комментариев** в коде (кроме godoc для экспортируемых сущностей)
- **Без лишнего экспорта** — только то, что нужно снаружи пакета
- **Error wrapping**: `fmt.Errorf("...: %w", err)`

## Наименование

- Файлы: `snake_case.go`
- Экспортируемые типы/функции: `PascalCase`
- Приватные поля/методы: `camelCase`
- Пакеты: `oneword`
- Slug: `kebab-case` (auto-generated из title если не указан)

## Структура файлов домена

```
internal/<domain>/
  model.go        — только GORM модель, теги json, gorm
  schema.go       — CreateInput, UpdateInput, Response DTO; теги validate
  repository.go   — интерфейс + реализация; GORM查询
  service.go      — бизнес-логика, вызов repository, валидация
  router.go       — Fiber handlers, route registration
  errors.go       — sentinel-ошибки, кастомные типы ошибок
```

## Валидация

- `go-playground/validator/v10` на схемах (CreateInput / UpdateInput)
- Проверка уникальности (email, slug) — в Service слое
- Ошибки валидации возвращать через `ValidationError` или маппинг полей

## Аутентификация и авторизация

- JWT access token в заголовке `Authorization: Bearer <token>`
- Middleware `auth.Required(roles...)` для защиты маршрутов
- User ID извлекается из токена и пробрасывается через `c.Locals("userID")`
- Роль извлекается через `c.Locals("role")`
- Ошибки: 401 Unauthorized (нет/невалидный токен), 403 Forbidden (не та роль)

## Работа с БД

- Все запросы через GORM
- Миграции только SQL-файлами
- Каскадное удаление для зависимых сущностей (ON DELETE CASCADE)
- Транзакции при submit теста (сохранение результата + ответов атомарно)

## HTTP

- JSON request/response
- Статус-коды: 200 OK, 201 Created, 400 Bad Request, 401 Unauthorized, 403 Forbidden, 404 Not Found, 409 Conflict, 500 Internal Server Error
- Ошибки: `{"error": "message"}` или `{"errors": {"field": "message"}}`
- Успешные пагинированные списки: `{"data": [...], "total": N, "page": P, "per_page": PP}`

## Добавление нового домена

1. Создать `internal/<domain>/` с model, schema, repository, service, router, errors
2. Добавить SQL миграцию
3. Зарегистрировать роуты в `RegisterRoutes()`
4. Связать DI в `bootstrap/app.go`
5. Написать тесты в `tests/<domain>/`

## Тестирование

- `package tests` (интеграционные, реальная БД)
- `helpers_test.go`: `setupTestApp()`, `getCleanDB()`, `seed<Entity>()`
- Cleanup: `TRUNCATE ... RESTART IDENTITY CASCADE` между тестами
- Три слоя: Repository → Service → Router
- `testify` (require/assert), Fiber `app.Test()`

## Вопросы и варианты ответов

- У вопроса ровно один правильный ответ (is_correct = true)
- Вопрос принадлежит либо тесту (`test_id`), либо уроку (`lesson_id`)
- При создании вопроса варианты передаются в одном запросе
- При получении вопроса (для прохождения) правильный ответ НЕ возвращается
- При получении вопроса (для редактирования) возвращается всё, включая правильный ответ
- Вопросы урока — для самопроверки: после submit сразу показывается правильный ответ

## Результаты теста

- Submit принимает `[{question_id, selected_option_id}]`
- Подсчёт: сравнение selected_option_id с правильным вариантом
- Всё сохраняется в одной транзакции (TestResult + TestAnswers)
- Grade = (score / max_score) * 100, округление до 2 знаков
- Результат возвращается сразу после submit

## Курсы и уроки

- Курс содержит упорядоченный список уроков (`sort_order`)
- Курс имеет цену (`price`); 0 — бесплатный курс
- Урок может содержать: контент (rich text), медиа (изображения/видео), ссылки на тесты, inline-вопросы
- `slug` курса и урока уникален и генерируется из title (transliteration + kebab-case)
- Уроки в курсе сортируются по `sort_order` ASC
- Reorder уроков: PATCH с массивом `[{id, sort_order}]`
- При удалении курса уроки удаляются каскадно

## Доступ к курсам

- Student видит только курсы, к которым есть доступ (фильтрация в Repository слое)
- Middleware `CourseAccess` проверяет доступ перед отдачей контента курса/урока
- Teacher/admin имеют доступ ко всем курсам без проверки
- Доступ выдаётся администратором (POST `/courses/:id/access`) или после оплаты
- При подтверждении платежа (status → `completed`) транзакционно создать CourseAccess
- Уникальность: один пользователь — одна запись доступа на курс (unique `(course_id, user_id)`)
- Курс считается доступным, если есть запись в `course_accesses` и `expires_at` > now() (или NULL)

## Платежи

- Создание платежа: POST `/courses/:id/payments` → статус `pending`
- Сумма фиксируется на момент создания из `courses.price`
- Только один pending payment на (user_id, course_id) (блокировать повторное создание)
- При смене статуса на `completed`:
  - Установить `paid_at = now()`
  - Создать CourseAccess (course_id, user_id, granted_by = NULL) в одной транзакции
- Статусы: `pending` → `completed` | `failed`; `completed` → `refunded` (при refund доступ не отзывается автоматически, только вручную)

## Slug

- У Course, Lesson, Page, Article slug обязателен и уникален в своей таблице
- Если slug не указан при создании — генерировать из title (транслитерация + kebab-case)
- При обновлении slug не менять (или менять только если явно передан)
