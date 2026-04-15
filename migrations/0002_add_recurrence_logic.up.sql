-- 1. Добавляем колонку для срока выполнения в основную таблицу tasks
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMPTZ;

-- 2. Создаем отдельную таблицу для хранения правил повторения (рекурсии)
CREATE TABLE IF NOT EXISTS task_recurrence_rules (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    
    type TEXT NOT NULL,          -- daily, monthly, parity, etc.
    every_n_days INT,           -- для ежедневных (каждый n-й день)
    days_of_month INT[],        -- для ежемесячных (числа месяца)
    specific_dates DATE[],      -- для конкретных дат
    parity TEXT,               -- even (четные), odd (нечетные)
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Индекс для быстрого поиска правил по ID задачи
CREATE INDEX IF NOT EXISTS idx_recurrence_task_id ON task_recurrence_rules (task_id);