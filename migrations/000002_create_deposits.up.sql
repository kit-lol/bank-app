-- Таблица с видами вкладов (Тарифы)
CREATE TABLE IF NOT EXISTS deposit_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    min_amount DECIMAL(15, 2) NOT NULL,
    interest_rate DECIMAL(5, 2) NOT NULL,
    can_deposit BOOLEAN DEFAULT FALSE,
    can_withdraw BOOLEAN DEFAULT FALSE
);

-- Таблица самих вкладов клиентов
CREATE TABLE IF NOT EXISTS deposits (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type_id INT NOT NULL REFERENCES deposit_types(id),
    amount DECIMAL(15, 2) NOT NULL,
    interest_rate DECIMAL(5, 2) NOT NULL,
    conditions JSONB,
    status VARCHAR(20) DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_accrual TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- История всех операций (пополнения, снятия)
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    deposit_id INT REFERENCES deposits(id) ON DELETE SET NULL,
    amount DECIMAL(15, 2) NOT NULL,
    operation_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексы для скорости
CREATE INDEX IF NOT EXISTS idx_deposits_user_id ON deposits(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);

-- === НАЧАЛЬНЫЕ ДАННЫЕ ===
-- Используем ON CONFLICT, чтобы не было ошибок, если данные уже есть
INSERT INTO deposit_types (id, name, description, min_amount, interest_rate, can_deposit, can_withdraw) VALUES
(1, 'TIDAL FLEX', 'Единственный вклад, который подстраивается под вас. Пополняйте и снимайте средства без ограничений.', 1000.00, 5.50, true, true),
(2, 'DEEP SAVER', 'Погрузитесь в мир высокой доходности. Пополняйте счет в любое время.', 5000.00, 7.00, true, false),
(3, 'CORAL FIX', 'Незыблемая стабильность и фиксированная прибыль. Идеальный выбор для тех, кто ценит предсказуемость.', 10000.00, 8.50, false, false),
(4, 'OCEAN ELITE', 'Максимальная доходность для избранных. Полная блокировка средств на весь срок.', 50000.00, 12.00, false, false),
(5, 'HARBOUR GOLD', 'Особые условия для клиентов, которые остаются с нами. Достойная ставка и надежность.', 25000.00, 9.50, true, true)
ON CONFLICT (id) DO NOTHING;

-- Сбрасываем счетчик ID
SELECT setval('deposit_types_id_seq', (SELECT MAX(id) FROM deposit_types));