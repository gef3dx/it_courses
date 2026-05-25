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
| **Course** | Курс: заголовок, описание, автор, список уроков |
| **Lesson** | Урок курса: контент, медиа (изображения, видео), порядок. К уроку можно прикреплять вопросы и тесты |
| **LessonMedia** | Медиафайл урока (image/video) |
| **LessonTestLink** | Связь урока с тестом |
| **CourseAccess** | Доступ пользователя к курсу: выдан администратором или после оплаты |
| **Payment** | Платёж пользователя за курс: сумма, статус, метод оплаты |
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

### Questions (тестовые)
| Метод | Путь | Роль | Описание |
|-------|------|------|----------|
| GET | `/tests/:id/questions` | any | Вопросы теста (без ответов) |
| POST | `/tests/:testId/questions` | teacher/admin | Создать вопрос + варианты |
| PUT | `/tests/:testId/questions/:id` | teacher/admin | Обновить вопрос |
| DELETE | `/tests/:testId/questions/:id` | teacher/admin | Удалить вопрос |

### Questions (урока)
| Метод | Путь | Роль | Описание |
|-------|------|------|----------|
| GET | `/lessons/:id/questions` | any | Вопросы урока (без ответов) |
| POST | `/lessons/:lessonId/questions` | teacher/admin | Создать вопрос урока + варианты |
| PUT | `/lessons/:lessonId/questions/:id` | teacher/admin | Обновить вопрос |
| DELETE | `/lessons/:lessonId/questions/:id` | teacher/admin | Удалить вопрос |
| POST | `/lessons/:id/submit` | student | Отправить ответы на вопросы урока |

### Test Results
| Метод | Путь | Роль | Описание |
|-------|------|------|----------|
| POST | `/tests/:id/submit` | student | Отправить ответы, получить результат |
| GET | `/results` | any | Мои результаты |
| GET | `/results/:id` | any | Детали результата |
| GET | `/tests/:id/results` | teacher/admin | Результаты всех студентов по тесту |

### Courses
| Метод | Путь | Роль | Описание |
|-------|------|------|----------|
| GET | `/courses` | any | Список доступных курсов (для student — только те, к которым есть доступ) |
| GET | `/courses/:slug` | any | Детали курса (с уроками; для student — только при наличии доступа) |
| POST | `/courses` | teacher/admin | Создать курс |
| PUT | `/courses/:id` | teacher/admin | Обновить курс |
| DELETE | `/courses/:id` | teacher/admin | Удалить курс |

### Course Access (доступ к курсам)
| Метод | Путь | Роль | Описание |
|-------|------|------|----------|
| GET | `/courses/:id/access` | admin | Список пользователей с доступом к курсу |
| POST | `/courses/:id/access` | admin | Выдать доступ пользователю `{user_id}` |
| DELETE | `/courses/:id/access/:userId` | admin | Отозвать доступ |
| GET | `/my/courses` | student | Мои курсы (к которым есть доступ) |

### Payments
| Метод | Путь | Роль | Описание |
|-------|------|------|----------|
| POST | `/courses/:id/payments` | student | Создать платёж за курс |
| GET | `/payments` | any | Мои платежи (student — свои, admin — все) |
| GET | `/payments/:id` | any | Детали платежа |
| PATCH | `/payments/:id/status` | admin | Обновить статус платежа (подтвердить/отменить) |

### Lessons
| Метод | Путь | Роль | Описание |
|-------|------|------|----------|
| GET | `/courses/:courseId/lessons` | any | Список уроков курса |
| GET | `/courses/:courseId/lessons/:id` | any | Детали урока (контент, медиа, тесты) |
| POST | `/courses/:courseId/lessons` | teacher/admin | Создать урок |
| PUT | `/courses/:courseId/lessons/:id` | teacher/admin | Обновить урок |
| DELETE | `/courses/:courseId/lessons/:id` | teacher/admin | Удалить урок |
| PATCH | `/courses/:courseId/lessons/reorder` | teacher/admin | Изменить порядок уроков |

### Lesson Test Links
| Метод | Путь | Роль | Описание |
|-------|------|------|----------|
| POST | `/lessons/:lessonId/tests` | teacher/admin | Привязать тест к уроку |
| DELETE | `/lessons/:lessonId/tests/:testId` | teacher/admin | Отвязать тест |

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

- **student** — просмотр курсов (только с доступом), уроков, статей, страниц; оплата курсов; прохождение тестов и вопросов урока, просмотр своих результатов
- **teacher** — создание/редактирование курсов, уроков, тестов, вопросов, статей, страниц; просмотр результатов студентов
- **admin** — всё + управление пользователями, выдача/отзыв доступа к курсам, управление платежами

## Доступ к курсам

- Студент видит в `/courses` только те курсы, на которые у него есть доступ
- Доступ выдаётся:
  - **Администратором** вручную (POST `/courses/:id/access`)
  - **Автоматически** после успешной оплаты (при смене статуса payment на `completed`)
- Доступ можно отозвать (DELETE `/courses/:id/access/:userId`)
- Список "Мои курсы" — GET `/my/courses`

## Оплата

1. Студент создаёт платёж (POST `/courses/:id/payments`) → статус `pending`
2. Администратор подтверждает оплату (PATCH `/payments/:id/status` → `completed`)
3. При подтверждении автоматически создаётся запись CourseAccess
4. Варианты статусов: `pending`, `completed`, `failed`, `refunded`

## Оценка теста

- Результат в процентах: `(правильные_ответы / всего_вопросов) * 100`
- Сохраняется `score`, `max_score`, `grade` (проценты)
- Возвращается сразу после отправки
