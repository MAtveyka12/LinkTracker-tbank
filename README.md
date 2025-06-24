# go-MAtveyka12

## Описание

**go-MAtveyka12** — это телеграм-бот и микросервис Scrapper для отслеживания изменений по ссылкам на GitHub-репозитории и вопросам Stack Overflow. Пользователь может добавлять ссылки для отслеживания, получать уведомления о новых коммитах, pull request'ах и ответах на вопросы.

---

## Стек технологий

- **Go 1.24**
- **PostgreSQL** — хранение данных
- **Redis** — кэширование
- **Kafka** — доставка событий между сервисами
- **Docker, Docker Compose** — контейнеризация и оркестрация
- **Liquibase** — миграции БД
- **Telebot** — работа с Telegram API

---

## Архитектура

Проект состоит из двух основных сервисов:
- **Bot** — Telegram-бот, взаимодействующий с пользователями, обрабатывающий команды и отправляющий уведомления.
- **Scrapper** — сервис, который отслеживает изменения по ссылкам (GitHub, Stack Overflow), сохраняет их в БД и отправляет события в Kafka или напрямую боту.

Взаимодействие между сервисами реализовано через Kafka (или HTTP), а данные пользователей и ссылок хранятся в PostgreSQL. Для ускорения работы используется Redis-кэш.

```
Пользователь <-> Telegram Bot <-> Scrapper <-> (GitHub/StackOverflow)
                        |           |
                      Redis      PostgreSQL
                        |
                      Kafka
```

---

## Основные понятия

- **Отслеживаемая ссылка** — пока что поддерживаются GitHub-репозитории и вопросы Stack Overflow.
- **Событие** — новое событие по ссылке (коммит, pull request, ответ на вопрос) отправляется пользователю.
- **Кэш** — для ускорения получения списка ссылок используется Redis.

---

## Структура репозитория

- `cmd/` — точки входа для сервисов (bot, scrapper)
- `internal/` — основная бизнес-логика и инфраструктура
- `config/` — конфигурационные файлы
- `migrations/` — миграции БД
- `templates/` — шаблоны сообщений для бота
- `patterns/` — паттерны для валидации ссылок
- `api/openapi/` — спецификации API

---

## Установка и запуск

### Быстрый старт через Docker Compose

```sh
docker-compose up --build
```

### Локальный запуск

1. Установите Go 1.24+ и PostgreSQL, Redis, Kafka.
2. Скопируйте `config/config.yaml` и создайте `.env` файл (см. ниже).
3. Соберите сервисы:
   ```sh
   make build
   ```
4. Запустите миграции (через liquibase или goose).
5. Запустите Scrapper:
   ```sh
   make run_scrapper
   ```
6. Запустите бота:
   ```sh
   make run_bot
   ```

---

## Переменные окружения

Создайте файл `.env` в корне проекта:

```env
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
GITHUB_TOKEN=your-github-token
DB_USER=db-user
DB_PASSWORD=db-password
KAFKA_BROKERS=port-kafka
KAFKA_TOPIC=kafka-topic
KAFKA_DLQ=dlq-topic
KAFKA_CONSUMER_GROUP=consumer-group
PATTERNS_FILE_PATH=patterns-path
TMPL_USER_DETAILS_FILE_PATH=user-details-path
TMPL_HELP_FILE_PATH=help-path
```

---

## Использование

- Запустите бота и найдите его в Telegram.
- Зарегистрируйтесь с помощью команды `/start`.
- Добавьте ссылку для отслеживания через `/track`.
- Получайте уведомления о новых событиях.

---

## Список команд бота

- `/start` — регистрация пользователя
- `/help` — список команд
- `/track` — начать отслеживание ссылки
- `/untrack` — прекратить отслеживание ссылки
- `/list` — показать список отслеживаемых ссылок

---

## Примеры API

### Регистрация чата
```
POST /tg-chat/{id}
```
Ответ: 200 OK — Чат зарегистрирован

### Получить все отслеживаемые ссылки
```
GET /links
Headers: Tg-Chat-Id: <id>
```
Ответ:
```json
{
  "links": [
    {"id": 1, "url": "https://github.com/user/repo", "tags": [], "filters": []}
  ],
  "size": 1
}
```

### Добавить ссылку
```
POST /links
Headers: Tg-Chat-Id: <id>
Body: { "link": "https://github.com/user/repo" }
```
Ответ: 200 OK — Ссылка успешно добавлена

### Удалить ссылку
```
DELETE /links
Headers: Tg-Chat-Id: <id>
Body: { "link": "https://github.com/user/repo" }
```
Ответ: 200 OK — Ссылка успешно убрана

---

## Тестирование

- Запуск всех тестов:
  ```sh
  make test
  ```
- Для интеграционных тестов используются контейнеры PostgreSQL, Kafka и Redis.

---