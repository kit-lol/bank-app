-- Создаем пользователя admin (пароль необходимо сменить при первом входе)
-- Хеш сгенерирован через bcrypt (cost=14)
INSERT INTO users (username, password_hash, balance, role, is_active) 
VALUES ('admin', '$2a$14$pOnlyjab7T5eZYY.z96qxudegDHOmrDmF7tM0XDjCWc/4r1ohZ/Z6', 0.00, 'admin', true)
ON CONFLICT (username) DO NOTHING;