# ── Финальный образ (только runtime, без Go) ─────────────────────────────────
# Бинарник собирается локально командой: go build -o main .
# и копируется в контейнер как уже готовый артефакт.
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Бинарник (собирается на хосте перед docker compose up)
COPY main .

# Шаблоны, статика, конфиг, миграции
COPY templates       ./templates
COPY static          ./static
COPY internal/config/config.yaml ./internal/config/config.yaml
COPY migrations      ./migrations

EXPOSE 8080

CMD ["./main"]