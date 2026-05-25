# AI Agent Instructions

Этот файл содержит контекст для AI-агентов, работающих над проектом **it_courses** — образовательная платформа API.

---

## Обязательное чтение перед работой

1. `specs/overview.md` — все сущности, маршруты, роли
2. `specs/architecture.md` — слои, DI, структура БД, JWT auth
3. `specs/guidelines.md` — код-стайл, правила, валидация

---

## Контекст для AI

Это Go-проект (Fiber v3 + GORM). Кодовая база уже имеет:
- Модуль `internal/user/` — референс для всех новых доменов
- Интеграционные тесты в `tests/user/` — референс для тестов
- Ручной DI в `internal/bootstrap/app.go`

---

## Правила

### Запрещено
- Менять README.md, specs/* без явного запроса
- Создавать файлы вне структуры (`internal/<domain>/`)
- Использовать AutoMigrate — только SQL-миграции
- Менять архитектуру (clean 3-layer) без согласования
- Добавлять emoji в код
- Комментировать код (кроме godoc экспортируемых сущностей)
- Редактировать go.mod/go.sum без необходимости

### Обязательно
- Следовать 3-layer архитектуре (Router → Service → Repository)
- Изучить существующий `internal/user/` перед кодингом нового домена
- Для нового домена: model, schema, repository, service, router, errors
- SQL миграция для новой таблицы
- DI в `bootstrap/app.go`
- JWT middleware для защищённых маршрутов
- Тесты в `tests/<domain>/` для нового домена
- `go build ./...` и `make test` перед завершением

### Submit теста
- Атомарная операция в транзакции
- Принимает массив `[{question_id, selected_option_id}]`
- Возвращает score, max_score, grade, правильные/неправильные ответы

### Submit вопросов урока
- Атомарная операция
- Принимает массив `[{question_id, selected_option_id}]`
- Возвращает score, max_score, grade + правильные ответы сразу

### Курсы и уроки
- Course 1→N Lesson (sort_order)
- Lesson может содержать: контент, медиа, тесты (lesson_test_links), вопросы (questions с lesson_id)
- Slug у курса и урока уникален
- Reorder: PATCH /courses/:courseId/lessons/reorder с массивом `[{id, sort_order}]`

### Доступ к курсам
- Student видит только курсы c доступом (check course_accesses)
- Teacher/admin — все курсы без проверки
- Middleware `CourseAccess` для защиты маршрутов курсов
- Доступ выдаётся: admin вручную или автоматически после оплаты

### Платежи
- POST `/courses/:id/payments` → status `pending`
- PATCH `/payments/:id/status` → `completed` → атомарно создать CourseAccess
- Статусы: pending → completed/failed, completed → refunded

### Email verification
- При регистрации: `email_verified_at = NULL`, генерировать UUID токен
- POST `/auth/verify-email` — подтверждает email по токену
- Login только при `email_verified_at IS NOT NULL`
- Интерфейс `EmailSender` для отправки (не реализовывать отправку, только интерфейс)

### Password recovery
- POST `/auth/forgot-password` → создаёт token (expires_at = now()+1h) → EmailSender
- POST `/auth/reset-password` → проверяет token → обновляет password_hash → удаляет token

### Self-edit и self-delete
- PUT `/users/:id` — middleware `OwnerOrAdmin` (user_id из JWT == :id или role == admin)
- DELETE `/users/:id` — self-delete требует `{password}` в body; admin без пароля

### Slug
- У Course, Lesson, Page, Article slug обязателен и уникален
- Если slug не указан при создании — генерировать из title (transliteration + kebab-case)
- При обновлении slug не менять (или менять только явно)

---

### Горутины и утечки
- Все блокирующие операции через `context.Context`
- `defer` для закрытия ресурсов (rows, body, timer, ticker)
- Не сохранять Fiber `c` в горутинах
- Каналы с явным жизненным циклом (отправитель/получатель)
- recover в каждой горутине

---

## Проверка перед завершением задачи

- [ ] `go build ./...` — компиляция без ошибок
- [ ] `make test` — все тесты проходят
- [ ] Новые маршруты зарегистрированы и защищены нужными middleware
- [ ] SQL миграция написана ( idempotent, с IF NOT EXISTS)
- [ ] DI в bootstrap/app.go обновлён
- [ ] Нет потенциальных утечек горутин (context, defer, recover)
