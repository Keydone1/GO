https://github.com/Keydone1/GO

## Структура
```
lab8/
─ go.mod
─ main.go
─ models.go
─ errors.go
─ validation.go
─ middleware.go
─ handlers.go
─ handlers_test.go
─ README.md
─ API.md
─ storage/
    ─ storage.go
    ─ memory.go
    ─ sqlite.go
```

## Требования
**Go 1.22+** (используются `r.PathValue("id")` и паттерны `"GET /users/{id}"`).

## Задания

### 1. Базовый REST CRUD
| Метод | Путь | Успех |
|-------|------|-------|
| GET | /users | 200 |
| GET | /users/{id} | 200 |
| POST | /users | 201 |
| PUT | /users/{id} | 200 |
| DELETE | /users/{id} | 204 |

### 2. База данных (SQLite)
`storage.UserStorage` — интерфейс. Реализации: `MemoryStorage` (для тестов) и `SQLiteStorage` (через `modernc.org/sqlite`, без CGO). `UNIQUE` на email → `ErrDuplicateEmail` → `409`.

### 3. Обработка ошибок и валидация
- `AppError{Code, Message, Details}` + `writeError` — единое место.
- `validation.go`: name (1..100), email (regexp, уникальный), age (0..150). Собирает **все** ошибки, не падает на первой.

### 4. Пагинация и фильтрация
Query: `name`, `min_age`, `max_age`, `page` (≥1), `limit` (1..100). Ответ: `{data, total, page, limit}`.

### 5. Unit-тесты
`handlers_test.go` — table-driven + `httptest.NewRecorder`. Покрыто: создание, дубликат, get, update, delete, пагинация/фильтры.

### 6. Документация
`API.md` — все маршруты, параметры, ошибки, примеры.

## Установка

```bash
cd lab8
go mod tidy
go build ./...
```

Если `go get modernc.org/sqlite` «висит»:
```bash
go env -w GOPROXY=https://goproxy.io,direct
go mod tidy
```

## Запуск

```bash
# SQLite (по умолчанию, файл users.db)
go run .

# In-memory
go run . -storage=memory

# Свои параметры
go run . -addr=:9090 -storage=sqlite -db=data.db
```

## Тестирование

### Автотесты
```bash
go test ./... -v
```
Ожидается: `PASS` для всех 6 тестов. `?   lab8/storage   [no test files]` — нормально.

С проверкой гонок:
```bash
go test ./... -race
```

### Через curl (PowerShell — используйте `curl.exe`)
```powershell
# Создать
curl.exe -X POST http://localhost:8080/users `
  -H "Content-Type: application/json" `
  -d "{\"name\":\"Alice\",\"email\":\"alice@example.com\",\"age\":30}"

# Список
curl.exe http://localhost:8080/users

# Фильтр + пагинация
curl.exe "http://localhost:8080/users?name=Al&min_age=18&page=1&limit=10"

# Один
curl.exe http://localhost:8080/users/1

# Обновить
curl.exe -X PUT http://localhost:8080/users/1 `
  -H "Content-Type: application/json" -d "{\"age\":31}"

# Удалить
curl.exe -X DELETE http://localhost:8080/users/1

# Ошибка валидации
curl.exe -X POST http://localhost:8080/users `
  -H "Content-Type: application/json" `
  -d "{\"name\":\"\",\"email\":\"bad\",\"age\":-1}"
```

> ⚠️ В PowerShell **не используйте** `-d "{\"..."}` — кавычки ломаются. Варианты:
> - `Invoke-RestMethod`:
>   ```powershell
>   Invoke-RestMethod -Uri http://localhost:8080/users -Method Post `
>     -ContentType "application/json" `
>     -Body (@{name="Alice"; email="alice@example.com"; age=30} | ConvertTo-Json)
>   ```
> - `curl.exe --data-raw '{\"...\"}'` (одинарные кавычки снаружи).
> - JSON в файле + `--data-binary "@body.json"`.

## Ожидаемые статусы

| Запрос | Код |
|--------|-----|
| POST валидный | 201 |
| POST невалидный | 400 |
| POST дубликат email | 409 |
| GET найден | 200 |
| GET не найден | 404 |
| GET невалидный id | 400 |
| PUT ок | 200 |
| DELETE ок | 204 |

## Частые проблемы

| Симптом | Решение |
|---------|---------|
| `is not in std` | `go mod init lab8` в корне, `go mod tidy` |
| `404` на `/` | Это норма — открывайте `/users` |
| `cannot use current (storage.User) as User` | Конвертация вручную в `Update` |
| `undefined: strconv` | Добавить `"strconv"` в импорты теста |
| `no new variables on left side of :=` | Переменная переименована — проверьте, что `updated` и `result` разные |

Документация API — в `API.md`.


package main

import (
    "encoding/json"
    "errors"
    "net/http"
)

// AppError — ошибка приложения, которую можно отдать клиенту.
type AppError struct {
    Code    int    `json:"-"`
    Message string `json:"error"`
    Details any    `json:"details,omitempty"`
}

func (e *AppError) Error() string { return e.Message }

func NewBadRequest(msg string, details any) *AppError {
    return &AppError{Code: http.StatusBadRequest, Message: msg, Details: details}
}

func NewNotFound(msg string) *AppError {
    return &AppError{Code: http.StatusNotFound, Message: msg}
}

func NewConflict(msg string) *AppError {
    return &AppError{Code: http.StatusConflict, Message: msg}
}

// writeError — единственное место, где ошибки превращаются в HTTP-ответы.
func writeError(w http.ResponseWriter, err error) {
    var appErr *AppError
    if !errors.As(err, &appErr) {
        appErr = &AppError{Code: http.StatusInternalServerError, Message: "internal server error"}
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(appErr.Code)
    json.NewEncoder(w).Encode(appErr)
}

// writeJSON — утилита для успешных ответов.
func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strconv"
    "testing"

    "lab8/storage"
)

func newTestMux(t *testing.T) *http.ServeMux {
    t.Helper()
    h := NewUserHandler(storage.NewMemoryStorage())
    mux := http.NewServeMux()
    mux.Handle("GET /users", APIHandler(h.List))
    mux.Handle("POST /users", APIHandler(h.Create))
    mux.Handle("GET /users/{id}", APIHandler(h.Get))
    mux.Handle("PUT /users/{id}", APIHandler(h.Update))
    mux.Handle("DELETE /users/{id}", APIHandler(h.Delete))
    return mux
}

func doJSON(t *testing.T, mux *http.ServeMux, method, path string, body any) *httptest.ResponseRecorder {
    t.Helper()
    var buf bytes.Buffer
    if body != nil {
        json.NewEncoder(&buf).Encode(body)
    }
    req := httptest.NewRequest(method, path, &buf)
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()
    mux.ServeHTTP(rec, req)
    return rec
}

func TestCreateUser(t *testing.T) {
    cases := []struct {
        name      string
        body      any
        wantCode  int
    }{
        {"valid", CreateUserRequest{Name: "Alice", Email: "a@x.com", Age: 30}, 201},
        {"empty name", CreateUserRequest{Name: "", Email: "a@x.com", Age: 30}, 400},
        {"bad email", CreateUserRequest{Name: "A", Email: "not-email", Age: 30}, 400},
        {"negative age", CreateUserRequest{Name: "A", Email: "a@x.com", Age: -1}, 400},
        {"age too big", CreateUserRequest{Name: "A", Email: "a@x.com", Age: 200}, 400},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            mux := newTestMux(t)
            rec := doJSON(t, mux, "POST", "/users", tc.body)
            if rec.Code != tc.wantCode {
                t.Errorf("got %d, want %d, body: %s", rec.Code, tc.wantCode, rec.Body.String())
            }
        })
    }
}

func TestCreateDuplicateEmail(t *testing.T) {
    mux := newTestMux(t)
    body := CreateUserRequest{Name: "A", Email: "a@x.com", Age: 20}
    if r := doJSON(t, mux, "POST", "/users", body); r.Code != 201 {
        t.Fatalf("first create: %d", r.Code)
    }
    if r := doJSON(t, mux, "POST", "/users", body); r.Code != 409 {
        t.Errorf("duplicate: got %d, want 409", r.Code)
    }
}

func TestGetUser(t *testing.T) {
    mux := newTestMux(t)
    rec := doJSON(t, mux, "POST", "/users", CreateUserRequest{Name: "A", Email: "a@x.com", Age: 20})
    var u User
    json.Unmarshal(rec.Body.Bytes(), &u)

    // найден
    if r := doJSON(t, mux, "GET", "/users/"+itoa(u.ID), nil); r.Code != 200 {
        t.Errorf("got %d, want 200", r.Code)
    }
    // не найден
    if r := doJSON(t, mux, "GET", "/users/9999", nil); r.Code != 404 {
        t.Errorf("got %d, want 404", r.Code)
    }
    // невалидный id
    if r := doJSON(t, mux, "GET", "/users/abc", nil); r.Code != 400 {
        t.Errorf("got %d, want 400", r.Code)
    }
}

func TestUpdateUser(t *testing.T) {
    mux := newTestMux(t)
    rec := doJSON(t, mux, "POST", "/users", CreateUserRequest{Name: "A", Email: "a@x.com", Age: 20})
    var u User
    json.Unmarshal(rec.Body.Bytes(), &u)

    newName := "Updated"
    upd := UpdateUserRequest{Name: &newName}
    r := doJSON(t, mux, "PUT", "/users/"+itoa(u.ID), upd)
    if r.Code != 200 {
        t.Fatalf("got %d, want 200", r.Code)
    }
    var got User
    json.Unmarshal(r.Body.Bytes(), &got)
    if got.Name != "Updated" || got.Age != 20 {
        t.Errorf("update wrong: %+v", got)
    }

    // невалидный возраст
    badAge := 999
    if r := doJSON(t, mux, "PUT", "/users/"+itoa(u.ID), UpdateUserRequest{Age: &badAge}); r.Code != 400 {
        t.Errorf("got %d, want 400", r.Code)
    }
}

func TestDeleteUser(t *testing.T) {
    mux := newTestMux(t)
    rec := doJSON(t, mux, "POST", "/users", CreateUserRequest{Name: "A", Email: "a@x.com", Age: 20})
    var u User
    json.Unmarshal(rec.Body.Bytes(), &u)

    if r := doJSON(t, mux, "DELETE", "/users/"+itoa(u.ID), nil); r.Code != 204 {
        t.Errorf("got %d, want 204", r.Code)
    }
    if r := doJSON(t, mux, "DELETE", "/users/"+itoa(u.ID), nil); r.Code != 404 {
        t.Errorf("got %d, want 404", r.Code)
    }
}

func TestListPaginationFilter(t *testing.T) {
    mux := newTestMux(t)
    seed := []CreateUserRequest{
        {"Alice", "a@x.com", 25},
        {"Bob", "b@x.com", 30},
        {"Alex", "c@x.com", 35},
        {"Charlie", "d@x.com", 40},
    }
    for _, s := range seed {
        doJSON(t, mux, "POST", "/users", s)
    }

    // Фильтр по имени
    r := doJSON(t, mux, "GET", "/users?name=Al", nil)
    var resp struct {
        Data  []User `json:"data"`
        Total int    `json:"total"`
    }
    json.Unmarshal(r.Body.Bytes(), &resp)
    if resp.Total != 2 {
        t.Errorf("name=Al: total=%d, want 2", resp.Total)
    }

    // Фильтр по возрасту
    r = doJSON(t, mux, "GET", "/users?min_age=30&max_age=40", nil)
    json.Unmarshal(r.Body.Bytes(), &resp)
    if resp.Total != 3 {
        t.Errorf("age 30..40: total=%d, want 3", resp.Total)
    }

    // Пагинация
    r = doJSON(t, mux, "GET", "/users?page=2&limit=2", nil)
    json.Unmarshal(r.Body.Bytes(), &resp)
    if resp.Total != 4 || len(resp.Data) != 2 {
        t.Errorf("page=2 limit=2: total=%d, len=%d", resp.Total, len(resp.Data))
    }
}

// мини-замена strconv.FormatInt, чтобы не импортировать
func itoa(i int64) string { return strconv.FormatInt(i, 10) }


package main

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"

    "lab8/storage"
)

// APIHandler — handler, возвращающий ошибку.
// ServeHTTP сам вызывает writeError при ошибке → централизованно.
type APIHandler func(w http.ResponseWriter, r *http.Request) error

func (h APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := h(w, r); err != nil {
        writeError(w, err)
    }
}

type UserHandler struct {
    storage storage.UserStorage
}

func NewUserHandler(s storage.UserStorage) *UserHandler {
    return &UserHandler{storage: s}
}

// ---------- GET /users ----------
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) error {
    f, err := parseFilter(r)
    if err != nil {
        return err
    }
    users, total, err := h.storage.List(r.Context(), f)
    if err != nil {
        return err
    }

    page := f.Offset/f.Limit + 1
    writeJSON(w, http.StatusOK, map[string]any{
        "data":  users,
        "total": total,
        "page":  page,
        "limit": f.Limit,
    })
    return nil
}

// ---------- GET /users/{id} ----------
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) error {
    id, err := parseID(r)
    if err != nil {
        return err
    }
    u, err := h.storage.Get(r.Context(), id)
    if errors.Is(err, storage.ErrNotFound) {
        return NewNotFound("user not found")
    }
    if err != nil {
        return err
    }
    writeJSON(w, http.StatusOK, u)
    return nil
}

// ---------- POST /users ----------
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) error {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return NewBadRequest("invalid JSON: "+err.Error(), nil)
    }
    if err := validateCreate(req); err != nil {
        return err
    }

    u, err := h.storage.Create(r.Context(), storage.User{
        Name: req.Name, Email: req.Email, Age: req.Age,
    })
    if errors.Is(err, storage.ErrDuplicateEmail) {
        return NewConflict("email already exists")
    }
    if err != nil {
        return err
    }
    writeJSON(w, http.StatusCreated, u)
    return nil
}

// ---------- PUT /users/{id} ----------
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) error {
    id, err := parseID(r)
    if err != nil {
        return err
    }

    current, err := h.storage.Get(r.Context(), id)
    if errors.Is(err, storage.ErrNotFound) {
        return NewNotFound("user not found")
    }
    if err != nil {
        return err
    }

    var req UpdateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return NewBadRequest("invalid JSON: "+err.Error(), nil)
    }

    // Конвертация storage.User → main.User для валидации
    currentAPI := User{
        ID:    current.ID,
        Name:  current.Name,
        Email: current.Email,
        Age:   current.Age,
    }

    // ← переменная называется updated (НЕ result!)
    updated, err := validateUpdate(req, currentAPI)
    if err != nil {
        return err
    }

    // ← переменная называется result
    result, err := h.storage.Update(r.Context(), id, storage.User{
        Name:  updated.Name,
        Email: updated.Email,
        Age:   updated.Age,
    })
    if errors.Is(err, storage.ErrDuplicateEmail) {
        return NewConflict("email already exists")
    }
    if err != nil {
        return err
    }

    writeJSON(w, http.StatusOK, result)
    return nil
}

// ---------- DELETE /users/{id} ----------
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) error {
    id, err := parseID(r)
    if err != nil {
        return err
    }
    if err := h.storage.Delete(r.Context(), id); err != nil {
        if errors.Is(err, storage.ErrNotFound) {
            return NewNotFound("user not found")
        }
        return err
    }
    w.WriteHeader(http.StatusNoContent)
    return nil
}

// ---------- helpers ----------

func parseID(r *http.Request) (int64, error) {
    s := r.PathValue("id")
    id, err := strconv.ParseInt(s, 10, 64)
    if err != nil || id <= 0 {
        return 0, NewBadRequest("invalid id", nil)
    }
    return id, nil
}

// parseFilter — пагинация и фильтрация из query-параметров.
func parseFilter(r *http.Request) (storage.Filter, error) {
    f := storage.Filter{Limit: 20, Offset: 0}
    q := r.URL.Query()

    if v := q.Get("name"); v != "" {
        f.NameLike = v
    }
    if v := q.Get("min_age"); v != "" {
        n, err := strconv.Atoi(v)
        if err != nil || n < 0 {
            return f, NewBadRequest("invalid min_age", nil)
        }
        f.MinAge = n
    }
    if v := q.Get("max_age"); v != "" {
        n, err := strconv.Atoi(v)
        if err != nil || n < 0 {
            return f, NewBadRequest("invalid max_age", nil)
        }
        f.MaxAge = n
    }

    page := 1
    if v := q.Get("page"); v != "" {
        n, err := strconv.Atoi(v)
        if err != nil || n < 1 {
            return f, NewBadRequest("invalid page", nil)
        }
        page = n
    }

    if v := q.Get("limit"); v != "" {
        n, err := strconv.Atoi(v)
        if err != nil || n < 1 || n > 100 {
            return f, NewBadRequest("invalid limit (1..100)", nil)
        }
        f.Limit = n
    }

    f.Offset = (page - 1) * f.Limit
    return f, nil
}

package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "lab8/storage"
)

func main() {
    var (
        addr        = flag.String("addr", ":8080", "server address")
        storageKind = flag.String("storage", "sqlite", "storage: memory | sqlite")
        dbPath      = flag.String("db", "users.db", "path to sqlite db")
    )
    flag.Parse()

    // ---- Хранилище ----
    var store storage.UserStorage
    switch *storageKind {
    case "memory":
        log.Println("Storage: in-memory")
        store = storage.NewMemoryStorage()
    case "sqlite":
        log.Println("Storage: sqlite, file =", *dbPath)
        s, err := storage.NewSQLiteStorage(*dbPath)
        if err != nil {
            log.Fatal("storage:", err)
        }
        store = s
    default:
        log.Fatalf("unknown storage: %s", *storageKind)
    }
    defer store.Close()

    // ---- Роутинг (Go 1.22+) ----
    h := NewUserHandler(store)
    mux := http.NewServeMux()
    mux.Handle("GET /users", APIHandler(h.List))
    mux.Handle("POST /users", APIHandler(h.Create))
    mux.Handle("GET /users/{id}", APIHandler(h.Get))
    mux.Handle("PUT /users/{id}", APIHandler(h.Update))
    mux.Handle("DELETE /users/{id}", APIHandler(h.Delete))

    // ---- Оборачиваем в middleware ----
    handler := loggingMiddleware(recoveryMiddleware(mux))

    srv := &http.Server{
        Addr:         *addr,
        Handler:      handler,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // ---- Graceful shutdown (из лабы 7) ----
    go func() {
        log.Printf("Listening on %s", *addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
    <-stop

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        fmt.Println("shutdown error:", err)
    }
    log.Println("Server stopped")
}

package main

import (
    "log"
    "net/http"
    "time"
)

// loggingMiddleware — логирует метод, URL и время выполнения.
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        rw := &statusRecorder{ResponseWriter: w, status: 200}
        next.ServeHTTP(rw, r)
        log.Printf("%s %s -> %d (%v)",
            r.Method, r.URL.RequestURI(), rw.status, time.Since(start))
    })
}

// recoveryMiddleware — ловит panic в handler'ах и отдаёт 500.
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if v := recover(); v != nil {
                log.Printf("PANIC: %v", v)
                writeError(w, &AppError{Code: 500, Message: "internal server error"})
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// statusRecorder запоминает HTTP-код ответа для логирования.
type statusRecorder struct {
    http.ResponseWriter
    status int
}

func (s *statusRecorder) WriteHeader(code int) {
    s.status = code
    s.ResponseWriter.WriteHeader(code)
}


package main

// User — модель пользователя (то, что отдаём клиенту).
type User struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

// CreateUserRequest — тело POST /users.
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

// UpdateUserRequest — тело PUT /users/{id}.
// Указатели, чтобы отличить "поле не передано" от "передано нулевое значение".
type UpdateUserRequest struct {
    Name  *string `json:"name,omitempty"`
    Email *string `json:"email,omitempty"`
    Age   *int    `json:"age,omitempty"`
}

package main

import (
    "fmt"
    "regexp"
    "strings"
)

var emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// validateCreate проверяет тело POST /users.
func validateCreate(req CreateUserRequest) error {
    errs := collectErrors(req.Name, req.Email, req.Age)
    if len(errs) > 0 {
        return NewBadRequest("validation failed", errs)
    }
    return nil
}

// validateUpdate проверяет тело PUT /users/{id}.
// Объединяем с текущим пользователем, чтобы получить итоговое состояние.
func validateUpdate(req UpdateUserRequest, current User) (User, error) {
    name := current.Name
    email := current.Email
    age := current.Age

    if req.Name != nil {
        name = *req.Name
    }
    if req.Email != nil {
        email = *req.Email
    }
    if req.Age != nil {
        age = *req.Age
    }

    errs := collectErrors(name, email, age)
    if len(errs) > 0 {
        return User{}, NewBadRequest("validation failed", errs)
    }
    return User{ID: current.ID, Name: name, Email: email, Age: age}, nil
}

func collectErrors(name, email string, age int) []string {
    var errs []string

    if strings.TrimSpace(name) == "" {
        errs = append(errs, "name is required")
    } else if len([]rune(name)) > 100 {
        errs = append(errs, fmt.Sprintf("name is too long (max 100 chars, got %d)", len([]rune(name))))
    }

    if !emailRe.MatchString(email) {
        errs = append(errs, "email is invalid")
    }

    if age < 0 {
        errs = append(errs, "age must be >= 0")
    } else if age > 150 {
        errs = append(errs, "age must be <= 150")
    }

    return errs
}


package storage

import (
    "context"
    "sort"
    "strings"
    "sync"
)

type MemoryStorage struct {
    mu     sync.RWMutex
    nextID int64
    users  map[int64]User
}

func NewMemoryStorage() *MemoryStorage {
    return &MemoryStorage{
        nextID: 1,
        users:  make(map[int64]User),
    }
}

func (m *MemoryStorage) List(_ context.Context, f Filter) ([]User, int, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    var filtered []User
    for _, u := range m.users {
        if f.NameLike != "" && !strings.Contains(strings.ToLower(u.Name), strings.ToLower(f.NameLike)) {
            continue
        }
        if f.MinAge > 0 && u.Age < f.MinAge {
            continue
        }
        if f.MaxAge > 0 && u.Age > f.MaxAge {
            continue
        }
        filtered = append(filtered, u)
    }

    sort.Slice(filtered, func(i, j int) bool { return filtered[i].ID < filtered[j].ID })

    total := len(filtered)

    // Пагинация
    if f.Offset >= len(filtered) {
        filtered = []User{}
    } else {
        filtered = filtered[f.Offset:]
        if f.Limit > 0 && f.Limit < len(filtered) {
            filtered = filtered[:f.Limit]
        }
    }

    return filtered, total, nil
}

func (m *MemoryStorage) Get(_ context.Context, id int64) (User, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    u, ok := m.users[id]
    if !ok {
        return User{}, ErrNotFound
    }
    return u, nil
}

func (m *MemoryStorage) Create(_ context.Context, u User) (User, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    // проверка уникальности email
    for _, existing := range m.users {
        if existing.Email == u.Email {
            return User{}, ErrDuplicateEmail
        }
    }
    u.ID = m.nextID
    m.nextID++
    m.users[u.ID] = u
    return u, nil
}

func (m *MemoryStorage) Update(_ context.Context, id int64, u User) (User, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    existing, ok := m.users[id]
    if !ok {
        return User{}, ErrNotFound
    }
    for otherID, other := range m.users {
        if otherID != id && other.Email == u.Email {
            return User{}, ErrDuplicateEmail
        }
    }
    u.ID = existing.ID
    m.users[id] = u
    return u, nil
}

func (m *MemoryStorage) Delete(_ context.Context, id int64) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    if _, ok := m.users[id]; !ok {
        return ErrNotFound
    }
    delete(m.users, id)
    return nil
}

func (m *MemoryStorage) Close() error { return nil }



package storage

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "strings"

    _ "modernc.org/sqlite"
)

type SQLiteStorage struct {
    db *sql.DB
}

func NewSQLiteStorage(path string) (*SQLiteStorage, error) {
    db, err := sql.Open("sqlite", path)
    if err != nil {
        return nil, fmt.Errorf("open db: %w", err)
    }
    // SQLite: одно соединение на запись — включаем WAL для параллелизма
    if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
        return nil, fmt.Errorf("pragma: %w", err)
    }

    schema := `
    CREATE TABLE IF NOT EXISTS users (
        id    INTEGER PRIMARY KEY AUTOINCREMENT,
        name  TEXT    NOT NULL,
        email TEXT    NOT NULL UNIQUE,
        age   INTEGER NOT NULL
    );`
    if _, err := db.Exec(schema); err != nil {
        return nil, fmt.Errorf("schema: %w", err)
    }

    return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) List(ctx context.Context, f Filter) ([]User, int, error) {
    // Динамически строим WHERE и args
    var conds []string
    var args []any

    if f.NameLike != "" {
        conds = append(conds, "name LIKE ?")
        args = append(args, "%"+f.NameLike+"%")
    }
    if f.MinAge > 0 {
        conds = append(conds, "age >= ?")
        args = append(args, f.MinAge)
    }
    if f.MaxAge > 0 {
        conds = append(conds, "age <= ?")
        args = append(args, f.MaxAge)
    }

    where := ""
    if len(conds) > 0 {
        where = " WHERE " + strings.Join(conds, " AND ")
    }

    // Всего записей для метаданных пагинации
    var total int
    if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users"+where, args...).Scan(&total); err != nil {
        return nil, 0, err
    }

    // Сама выборка
    query := "SELECT id, name, email, age FROM users" + where + " ORDER BY id"
    queryArgs := append([]any{}, args...)
    if f.Limit > 0 {
        query += " LIMIT ? OFFSET ?"
        queryArgs = append(queryArgs, f.Limit, f.Offset)
    }

    rows, err := s.db.QueryContext(ctx, query, queryArgs...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Age); err != nil {
            return nil, 0, err
        }
        users = append(users, u)
    }
    return users, total, rows.Err()
}

func (s *SQLiteStorage) Get(ctx context.Context, id int64) (User, error) {
    var u User
    err := s.db.QueryRowContext(ctx,
        "SELECT id, name, email, age FROM users WHERE id = ?", id).
        Scan(&u.ID, &u.Name, &u.Email, &u.Age)
    if errors.Is(err, sql.ErrNoRows) {
        return User{}, ErrNotFound
    }
    return u, err
}

func (s *SQLiteStorage) Create(ctx context.Context, u User) (User, error) {
    res, err := s.db.ExecContext(ctx,
        "INSERT INTO users(name, email, age) VALUES(?, ?, ?)",
        u.Name, u.Email, u.Age)
    if err != nil {
        if isUniqueErr(err) {
            return User{}, ErrDuplicateEmail
        }
        return User{}, err
    }
    id, _ := res.LastInsertId()
    u.ID = id
    return u, nil
}

func (s *SQLiteStorage) Update(ctx context.Context, id int64, u User) (User, error) {
    res, err := s.db.ExecContext(ctx,
        "UPDATE users SET name=?, email=?, age=? WHERE id=?",
        u.Name, u.Email, u.Age, id)
    if err != nil {
        if isUniqueErr(err) {
            return User{}, ErrDuplicateEmail
        }
        return User{}, err
    }
    n, _ := res.RowsAffected()
    if n == 0 {
        return User{}, ErrNotFound
    }
    u.ID = id
    return u, nil
}

func (s *SQLiteStorage) Delete(ctx context.Context, id int64) error {
    res, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
    if err != nil {
        return err
    }
    n, _ := res.RowsAffected()
    if n == 0 {
        return ErrNotFound
    }
    return nil
}

func (s *SQLiteStorage) Close() error { return s.db.Close() }

func isUniqueErr(err error) bool {
    return strings.Contains(strings.ToLower(err.Error()), "unique")
}


package storage

import (
    "context"
    "errors"
)

// User — модель на уровне хранилища (может совпадать с моделью API).
type User struct {
    ID    int64
    Name  string
    Email string
    Age   int
}

// Filter — параметры фильтрации и пагинации.
type Filter struct {
    NameLike string // LIKE %NameLike%
    MinAge   int    // 0 = не фильтровать
    MaxAge   int    // 0 = не фильтровать
    Limit    int    // 0 = не ограничивать
    Offset   int
}

// Ошибки хранилища — не HTTP, а «бизнес-ошибки».
// Handlers преобразуют их в HTTP-коды.
var (
    ErrNotFound       = errors.New("user not found")
    ErrDuplicateEmail = errors.New("email already exists")
)

// UserStorage — абстракция над БД.
// Позволяет переключать Memory <-> SQLite без изменения handlers.
type UserStorage interface {
    List(ctx context.Context, f Filter) (users []User, total int, err error)
    Get(ctx context.Context, id int64) (User, error)
    Create(ctx context.Context, u User) (User, error)
    Update(ctx context.Context, id int64, u User) (User, error)
    Delete(ctx context.Context, id int64) error
    Close() error
}