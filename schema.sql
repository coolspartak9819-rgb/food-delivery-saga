-- schema.sql
-- Этот скрипт создает таблицы для хранения состояния заказов и логов выполнения Саги.

-- Таблица для хранения заказов
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY,
    status VARCHAR(50) NOT NULL,
    items JSONB, -- Используем JSONB для гибкого хранения состава заказа
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Журнал выполнения шагов Саги. Это сердце отказоустойчивости.
-- Перед выполнением шага мы пишем сюда лог со статусом 'STARTED'.
-- После успеха обновляем на 'SUCCESS'.
-- При откате обновляем на 'COMPENSATED'.
CREATE TABLE IF NOT EXISTS saga_logs (
    id BIGSERIAL PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    step_name VARCHAR(100) NOT NULL,
    step_status VARCHAR(50) NOT NULL, -- e.g., 'STARTED', 'SUCCESS', 'FAILED', 'COMPENSATED'
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Индекс для быстрого поиска логов по конкретному заказу
CREATE INDEX IF NOT EXISTS idx_saga_logs_order_id ON saga_logs (order_id);
