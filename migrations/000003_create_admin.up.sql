-- Создаем пользователя admin (пароль необходимо сменить при первом входе)
-- Хеш сгенерирован через bcrypt (cost=10)
INSERT INTO users (username, password_hash, balance, role, is_active) 
VALUES ('admin', '$2a$10$27NTm6UW5oAaI98TrxYP5eTbp0iqgGzPO0hC9c29SLOIvjD1S8/aS', 0.00, 'admin', true)
ON CONFLICT (username) DO NOTHING;