-- Добавляем недостающие поля user-модуля в уже существующую таблицу users.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS phone TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS first_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS last_name TEXT NOT NULL DEFAULT '';

-- Индекс по телефону пригодится для поиска и проверки уникальности в будущем.
CREATE INDEX IF NOT EXISTS idx_users_phone ON users (phone);
