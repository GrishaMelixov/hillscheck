<div align="center">

# 💰 WealthCheck

### Fintech-платформа личных финансов с RPG-геймификацией

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-316192?style=for-the-badge&logo=postgresql&logoColor=white)](https://postgresql.org)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io)
[![React](https://img.shields.io/badge/React-18-20232A?style=for-the-badge&logo=react&logoColor=61DAFB)](https://react.dev)
[![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://docker.com)

</div>

---

Каждая трата прокачивает персонажа. Ресторан снижает ❤️ HP, спортзал повышает 💪 Strength, книги — 🧠 Intellect, развлечения — 🔮 Mana. Загрузи выписку из банка — и посмотри на свои финансы по-новому.

---

## ✨ Возможности

| | Фича | Описание |
|---|---|---|
| 🎮 | **RPG-геймификация** | Каждая транзакция влияет на атрибуты персонажа — HP, Mana, Strength, Intellect, Luck |
| 🤖 | **AI-классификация** | Гибридный подход: MCC-таблица (90+ категорий) → Gemini 2.0 Flash → Ollama llama3.2 → fallback |
| 📸 | **OCR скриншотов** | Tesseract офлайн — парсит скриншоты T-Bank, Сбер, кассовые чеки без внешних API |
| ⚡ | **Real-time обновления** | WebSocket: после обработки транзакции профиль обновляется мгновенно |
| 🔁 | **Идемпотентный импорт** | Повторная загрузка того же CSV безопасна — `ON CONFLICT DO NOTHING` + суммы в `int64` |
| 📊 | **Аналитика с кешем** | Redis TTL 1 час + составные индексы PostgreSQL — ускорение запросов ~4x |
| 🔐 | **JWT + Refresh tokens** | Refresh-токены в Redis, access — stateless JWT |
| 🐳 | **Одна команда** | `docker compose up --build` — полный стек за минуту |

---

## 🏗 Архитектура

```mermaid
flowchart TD
    A[React 18 SPA\nTypeScript + Tailwind] -->|HTTP / WebSocket| B[Go HTTP Server\nchi v5]

    B --> C[JWT Middleware]
    C --> D[Handlers\ninternal/adapter/http]

    D --> E[Usecase Layer\nбизнес-логика]

    E --> F[PostgreSQL 16\npgx/v5 pool]
    E --> G[Redis 7\ngo-redis/v9]
    E --> H[Worker Pool\nкастомный горутин-пул]

    H --> I{AI Classifier\nChainedProvider}
    I -->|1. MCC table| J[ISO 18245\n90+ категорий]
    I -->|2. LLM| K[Gemini 2.0 Flash\n/ Ollama llama3.2]
    I -->|3. fallback| L[Static rules]

    H --> M[Game Engine\nRPG атрибуты]
    M --> N[WebSocket Hub\ngorilla/websocket]
    N --> A

    O[CSV / Скриншот] -->|import| D
    O2[Tesseract OCR\nрусский язык] -->|парсинг| O

    style A fill:#1e293b,color:#e2e8f0
    style M fill:#1b4332,color:#ffffff
    style I fill:#0f4c75,color:#ffffff
    style N fill:#4a1942,color:#ffffff
```

---

## 🎮 Как работает RPG-механика

```
Транзакция → AI-классификация → Game Engine → WebSocket push → обновление профиля

Ресторан / фастфуд  →  ❤️  HP      −5
Спортзал / спорт    →  💪  Strength +3
Книги / курсы       →  🧠  Intellect +4
Развлечения         →  🔮  Mana     −3
Путешествия         →  🍀  Luck     +2
```

Каждое изменение атрибута логируется в `game_events` с привязкой к транзакции — полная история прокачки.

---

## 🤖 Гибридная классификация транзакций

```
Входящая транзакция
        │
        ▼
┌───────────────────┐
│  MCC-таблица      │  ← ISO 18245, 90+ категорий, детерминировано
│  (слой 1)         │     попадание → 100% точность, 0 мс overhead
└────────┬──────────┘
    miss │
        ▼
┌───────────────────┐
│  Gemini 2.0 Flash │  ← LLM fallback для нестандартных MCC
│  (слой 2)         │
└────────┬──────────┘
    fail │
        ▼
┌───────────────────┐
│  Ollama llama3.2  │  ← локальный LLM, без квот и внешних зависимостей
│  (слой 3)         │
└────────┬──────────┘
    fail │
        ▼
┌───────────────────┐
│  Static fallback  │  ← никогда не блокирует pipeline
└───────────────────┘
```

---

## ⚡ Ключевые технические решения

**Идемпотентный импорт**
```sql
INSERT INTO transactions (account_id, external_id, amount, ...)
VALUES (...)
ON CONFLICT (account_id, external_id) DO NOTHING;
-- Повторная загрузка того же CSV — безопасна
-- Суммы в int64 (копейки) — никаких float-ошибок
```

**Redis-кеш аналитики**
```
Cache key: analytics:{account_id}:{period_days}
TTL: 1 час
Cache miss → PostgreSQL
  INDEX (account_id, occurred_at DESC)
  INDEX (account_id, clean_category)
Ускорение: ~4x
```

**OCR без внешних API**
```
Tesseract + ru.traineddata → встроен в Docker-образ
Парсит: T-Bank, Сбербанк, кассовые чеки
Полностью офлайн, без квот, бесплатно
```

---

## 🚀 Быстрый старт

```bash
git clone https://github.com/GrishaMelixov/WealthCheck
cd WealthCheck

cp .env.example .env
# Заполнить JWT_SECRET и пароли БД

docker compose up --build
# → http://localhost:8080
```

---

## ⚙️ Конфигурация

| Переменная | Описание | Default |
|---|---|---|
| `DATABASE_URL` | PostgreSQL DSN | — |
| `REDIS_URL` | Redis DSN | — |
| `JWT_SECRET` | 32+ байт случайная строка | — |
| `PROVIDER` | `gemini` / `ollama` / `chained` | `chained` |
| `GEMINI_API_KEY` | Опционально, для AI-классификации | — |
| `GEMINI_MODEL` | Модель Gemini | `gemini-2.0-flash` |
| `WORKER_COUNT` | Размер горутин-пула | `10` |
| `QUEUE_SIZE` | Размер очереди задач | `500` |

---

## 🌐 API

```
POST  /auth/register
POST  /auth/login
POST  /auth/refresh
POST  /auth/logout

GET   /api/v1/accounts
POST  /api/v1/transactions/import        # CSV (Tinkoff / Сбер формат)
GET   /api/v1/transactions?account_id=&limit=&offset=
POST  /api/v1/receipts/parse             # скриншот → транзакции (Tesseract OCR)
GET   /api/v1/analytics/summary?account_id=&days=30
GET   /api/v1/profile                    # RPG-профиль персонажа
GET   /api/v1/quests                     # активные квесты
GET   /ws?token=                         # WebSocket (real-time пуши)
GET   /health
```

---

## 🧪 Тестирование

```bash
# Unit-тесты (~87% покрытие core business logic)
go test ./internal/usecase/...

# Интеграционные тесты (testcontainers-go — реальный PostgreSQL в Docker)
go test -tags=integration ./internal/integration/... -v
```

Покрыты юнит-тестами: `game_engine`, `transaction_import`, `register/login/logout/refresh`, `get_quests`, `get_profile`

Покрыты интеграционными: идемпотентность импорта, обновление `game_profiles` через engine, аналитические запросы

---

## 📁 Структура репозитория

```
WealthCheck/
├── cmd/server/               # Точка входа, DI вручную
├── internal/
│   ├── domain/               # Сущности: User, Account, Transaction, GameProfile, GameEvent
│   ├── usecase/              # Бизнес-логика (без зависимостей от фреймворков)
│   │   └── port/             # Интерфейсы репозиториев и провайдеров
│   ├── adapter/
│   │   ├── http/             # Хендлеры, JWT middleware, rate limit
│   │   ├── postgres/         # Реализация репозиториев (pgx/v5)
│   │   ├── ai/               # Gemini, Ollama, Chained, MCC-таблица
│   │   └── websocket/        # WebSocket hub (gorilla/websocket)
│   ├── infrastructure/
│   │   ├── auth/             # JWT сервис, Redis token store
│   │   ├── cache/            # Redis подключение
│   │   └── worker/           # Горутин-пул для async обработки
│   └── integration/          # Интеграционные тесты (testcontainers-go)
├── migrations/               # 7 SQL-миграций (embedded через go:embed)
├── frontend/                 # React 18 + TypeScript + Tailwind v3
├── Dockerfile                # 3-stage: Node → Go → Alpine+Tesseract
├── docker-compose.yml
└── .env.example
```

---

## 🛠 Стек

**Backend:** Go 1.25 · chi v5 · pgx/v5 · go-redis/v9 · gorilla/websocket · golang-jwt · golang-migrate · uber-go/zap · testcontainers-go

**AI/ML:** Gemini 2.0 Flash · Ollama llama3.2 · Tesseract OCR (русский)

**Frontend:** React 18 · TypeScript · Vite · Tailwind CSS v3 · iOS 26 Liquid Glass design

**Infrastructure:** Docker · Docker Compose · PostgreSQL 16 · Redis 7

---

<div align="center">

Сделано с 🖤 на Go · [MIT License](LICENSE)

</div>
