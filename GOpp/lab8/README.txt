# Lab 8 — Users REST API

REST API для управления пользователями на Go: полный CRUD, хранение в SQLite (или in-memory), валидация, централизованная обработка ошибок, пагинация и фильтрация, unit-тесты.

---

## Содержание

- [Возможности](#возможности)
- [Требования](#требования)
- [Установка](#установка)
- [Запуск](#запуск)
- [Тесты](#тесты)
- [Маршруты API](#маршруты-api)
- [Примеры запросов](#примеры-запросов)
- [Структура проекта](#структура-проекта)
- [Архитектура](#архитектура)
- [Документация API](#документация-api)
- [Решение проблем](#решение-проблем)

---

## Возможности

- ✅ Полный CRUD для пользователей (`GET`, `POST`, `PUT`, `DELETE`)
- ✅ Хранение в **SQLite** (файл) или **in-memory** (без БД)
- ✅ **Валидация** входных данных (имя, email, возраст)
- ✅ **Централизованная обработка ошибок** — единый JSON-формат
- ✅ **Пагинация** (`page`, `limit`)
- ✅ **Фильтрация** по имени и возрасту (`name`, `min_age`, `max_age`)
- ✅ **Уникальность email** — возвращает `409 Conflict`
- ✅ **Middleware**: логирование запросов + восстановление после `panic`
- ✅ **Graceful shutdown** по `Ctrl+C`
- ✅ **Unit-тесты** для всех маршрутов
- ✅ Роутинг на **Go 1.22+** (`GET /users/{id}` без внешних библиотек)

---

## Требования

| Компонент | Версия | Зачем |
|-----------|--------|-------|
| **Go** | **1.22+** | Нужен `r.PathValue()` и паттерны `"GET /users/{id}"` |
| **SQLite** | встроен | Через `modernc.org/sqlite` — **не требует** установки сервера БД или CGO |

### Проверка версии Go

```bash
go version
# ожидается: go version go1.22.x ...
```

Если версия ниже 1.22 — обновите Go: https://go.dev/dl/

---

## Установка

```bash
# 1. Перейти в папку проекта
cd C:\labs\GOpp\lab8

# 2. Скачать зависимости
go mod download

# 3. Проверить, что всё собирается
go build ./...
```

Если видите ошибку `missing go.sum entry` — выполните `go mod tidy` и повторите `go build ./...`.

---

## Запуск

### Вариант 1: SQLite (по умолчанию)

```bash
go run .
```

Сервер на `http://localhost:8080`, БД в файле `users.db` (создаётся автоматически).

### Вариант 2: In-memory (без файла БД)

```bash
go run . -storage=memory
```

Данные живут только в памяти — **пропадают при перезапуске**. Удобно для демо и тестов.

### Вариант 3: Свои параметры

```bash
go run . -addr=:9090 -storage=sqlite -db=data/prod.db
```

| Флаг | По умолчанию | Описание |
|------|--------------|----------|
| `-addr` | `:8080` | Адрес и порт сервера |
| `-storage` | `sqlite` | `memory` или `sqlite` |
| `-db` | `users.db` | Путь к файлу SQLite |

### Сборка в бинарник

```bash
go build -o lab8.exe
./lab8.exe -storage=memory
```

### Остановка

Нажмите `Ctrl+C` в терминале — сервер корректно завершит работу (graceful shutdown, до 10 секунд на завершение активных запросов).

---

## Тесты

```bash
go test ./... -v
```

Ожидаемый вывод:

```
=== RUN   TestCreateUser
=== RUN   TestCreateUser/valid
=== RUN   TestCreateUser/empty_name
=== RUN   TestCreateUser/bad_email
=== RUN   TestCreateUser/negative_age
=== RUN   TestCreateUser/age_too_big
--- PASS: TestCreateUser (0.00s)
=== RUN   TestCreateDuplicateEmail
--- PASS: TestCreateDuplicateEmail (0.00s)
=== RUN   TestGetUser
--- PASS: TestGetUser (0.00s)
=== RUN   TestUpdateUser
--- PASS: TestUpdateUser (0.00s)
=== RUN   TestDeleteUser
--- PASS: TestDeleteUser (0.00s)
=== RUN   TestListPaginationFilter
--- PASS: TestListPaginationFilter (0.00s)
PASS
ok      lab8    0.012s
?       lab8/storage    [no test files]
```

Строка `?   lab8/storage   [no test files]` — нормально: в пакете `storage` нет `_test.go`.

### Полезные флаги

```bash
go test ./... -v -run TestCreateUser   # только один тест
go test ./... -cover                    # покрытие в %
go test ./... -race                     # проверка гонок данных
```

---

## Маршруты API

Базовый URL: `http://localhost:8080`

| Метод | Путь | Описание | Успех |
|-------|------|----------|-------|
| `GET` | `/users` | Список с пагинацией и фильтрацией | `200` |
| `GET` | `/users/{id}` | Один пользователь | `200` |
| `POST` | `/users` | Создать пользователя | `201` |
| `PUT` | `/users/{id}` | Обновить (частично) | `200` |
| `DELETE` | `/users/{id}` | Удалить | `204` |

### Модель `User`

```json
{
  "id": 1,
  "name": "Alice",
  "email": "alice@example.com",
  "age": 30
}
```

| Поле | Тип | Ограничения |
|------|-----|-------------|
| `id` | int64 | Генерируется сервером |
| `name` | string | 1..100 символов |
| `email` | string | Валидный email, **уникальный** |
| `age` | int | 0..150 |

### Query-параметры для `GET /users`

| Параметр | Тип | По умолчанию | Описание |
|----------|-----|--------------|----------|
| `name` | string | — | Подстрока имени (регистронезависимо) |
| `min_age` | int | — | Минимальный возраст включительно |
| `max_age` | int | — | Максимальный возраст включительно |
| `page` | int | `1` | Номер страницы (1-based) |
| `limit` | int | `20` | Размер страницы (1..100) |

---

## Примеры запросов

### Создать пользователя

**PowerShell:**

```powershell
curl.exe -X POST http://localhost:8080/users `
  -H "Content-Type: application/json" `
  -d "{\"name\":\"Alice\",\"email\":\"alice@example.com\",\"age\":30}"
```

**bash / Git Bash:**

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","age":30}'
```

Ответ `201 Created`:

```json
{"id": 1, "name": "Alice", "email": "alice@example.com", "age": 30}
```

### Список пользователей

```bash
curl http://localhost:8080/users
```

Ответ `200 OK`:

```json
{
  "data": [{"id": 1, "name": "Alice", "email": "alice@example.com", "age": 30}],
  "total": 1,
  "page": 1,
  "limit": 20
}
```

### С фильтром и пагинацией

```bash
curl "http://localhost:8080/users?name=Ali&min_age=18&page=1&limit=10"
```

### Один пользователь

```bash
curl http://localhost:8080/users/1
```

### Обновить (только переданные поля)

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"age":31}'
```

### Удалить

```bash
curl -X DELETE http://localhost:8080/users/1
# → 204 No Content
```

### Ошибка валидации

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"","email":"bad","age":-1}'
```

Ответ `400 Bad Request`:

```json
{
  "error": "validation failed",
  "details": ["name is required", "email is invalid", "age must be >= 0"]
}
```

### Дубликат email

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Bob","email":"alice@example.com","age":25}'
```

Ответ `409 Conflict`:

```json
{"error": "email already exists"}
```

---

## Структура проекта

```
lab8/
├── go.mod               # модуль и зависимости
├── go.sum               # контрольные суммы зависимостей
├── README.md            # этот файл
├── API.md               # документация API
├── main.go              # точка входа, роутинг, graceful shutdown
├── models.go            # User, CreateUserRequest, UpdateUserRequest
├── errors.go            # AppError, writeError, writeJSON
├── validation.go        # validateCreate, validateUpdate
├── middleware.go        # loggingMiddleware, recoveryMiddleware
├── handlers.go          # HTTP-обработчики (List, Get, Create, Update, Delete)
├── handlers_test.go     # unit-тесты
└── storage/
    ├── storage.go       # интерфейс UserStorage + Filter + ошибки
    ├── memory.go        # in-memory реализация
    └── sqlite.go        # SQLite реализация
```

---

## Архитектура

### Слои

```
┌─────────────────────────────────┐
│         HTTP (net/http)         │  ← handlers.go
│   Роутинг, JSON, коды ответов   │
└───────────────┬─────────────────┘
                │
┌───────────────▼─────────────────┐
│      Бизнес-логика              │  ← validation.go, errors.go
│   Валидация, преобразование     │
│   ошибок в HTTP-коды            │
└───────────────┬─────────────────┘
                │
┌───────────────▼─────────────────┐
│     UserStorage (интерфейс)     │  ← storage/storage.go
│     List/Get/Create/Update/     │
│     Delete                      │
└───────┬─────────────────┬───────┘
        │                 │
┌───────▼──────┐  ┌───────▼───────┐
│ MemoryStorage│  │ SQLiteStorage │
└──────────────┘  └───────────────┘
```

### Ключевые решения

- **`UserStorage` — интерфейс.** Handlers не знают, память там или SQLite. Переключение — через флаг `-storage`.
- **`APIHandler`** — handler, возвращающий `error`. `ServeHTTP` вызывает `writeError` — **единое место** превращения ошибок в HTTP-ответы.
- **`UpdateUserRequest` с указателями** — отличает «поле не передали» (`nil`) от «передали пустое значение» (`&""`).
- **`sync.RWMutex`** в `MemoryStorage` — HTTP-сервер многопоточный, `map` не потокобезопасна.
- **SQLite через `modernc.org/sqlite`** — чистый Go, без CGO и без установки сервера.
- **`PRAGMA journal_mode=WAL`** — читатели не блокируют писателей.
- **`?`-плейсхолдеры** — защита от SQL-инъекций.

---

## Документация API

Подробная документация со всеми параметрами, кодами ответов и примерами — в файле **[API.md](API.md)**.

Быстрый просмотр в VS Code:

1. Откройте `API.md`.
2. Нажмите `Ctrl+Shift+V` — откроется отрендеренная версия.
3. Для бокового просмотра — `Ctrl+K V`.

---

## Решение проблем

### `go: cannot find main module`

Вы не в папке проекта. Перейдите в корень:

```bash
cd C:\labs\GOpp\lab8
```

Там должен лежать `go.mod`.

### `package lab8/... is not in std`

Импорт с префиксом, не совпадающим с `module` в `go.mod`. Проверьте первую строку `go.mod` — она должна быть `module lab8`.

### `bind: address already in use`

Порт `:8080` занят. Смените:

```bash
go run . -addr=:9090
```

Или найдите и остановите процесс:

```powershell
# Windows
netstat -ano | findstr :8080
taskkill /PID <pid> /F
```

### `missing go.sum entry`

```bash
go mod tidy
go mod download
```

### `no new variables on left side of :=`

Где-то используется `:=` для уже объявленных переменных. Замените на `=` или переименуйте одну из переменных. Подробнее — в разделе «Ошибка `no new variables`».

### `cannot use current (storage.User) as User value`

В проекте **два типа `User`**: в `main` и в `storage`. В `handlers.go` нужна явная конвертация:

```go
currentAPI := User{
    ID:    current.ID,
    Name:  current.Name,
    Email: current.Email,
    Age:   current.Age,
}
```

### `undefined: strconv` в тестах

В `handlers_test.go` добавьте `"strconv"` в блок `import`.

### SQLite `database is locked`

Обычно из-за параллельной записи. Убедитесь, что включён WAL (`PRAGMA journal_mode=WAL` — уже в коде). Если проблема остаётся — не запускайте два экземпляра сервера на одном файле `users.db`.

### `-race` показывает предупреждения

Запустите с флагом:

```bash
go test ./... -race
go run -race .
```

Исправьте найденные гонки (обычно — добавить `Lock/Unlock` вокруг общих данных).

---

## Лицензия

Учебный проект. Свободно используйте и модифицируйте.

---

## Автор

Студент группы `<номер группы>`  
`<ФИО>`  
Лабораторная работа №8, 2024/2025 уч. год