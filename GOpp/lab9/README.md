https://github.com/Keydone1/GO

## Структура
```
lab9/
├── go.mod
├── server/
│   ├── main.go
│   ├── models.go
│   ├── errors.go
│   ├── validation.go
│   ├── middleware.go
│   ├── handlers.go
│   ├── auth.go
│   └── storage/
│       ├── storage.go
│       ├── memory.go
│       └── sqlite.go
└── client/
    └── main.go
```

## Требования
**Go 1.22+**. Зависимости: `golang.org/x/crypto` (bcrypt), `modernc.org/sqlite`.

## Задания

### 1. Клиент + CRUD-меню
`client/main.go` — структура `Client`, все CRUD-операции через `doRequest`.

### 2. Обработка ответов и ошибок
- `doRequest` — единая точка, добавляет `Authorization: Bearer <token>`.
- `parseAPIError` — превращает `{"error":..., "details":[...]}` в `*APIError`.
- `printUsersTable`, `printUser` — форматированный вывод.

### 3. Консольный интерфейс
Два меню:
- **гость**: Регистрация / Вход / Выход.
- **авторизован**: Список / Один / Создать / Обновить / Удалить / Выйти из аккаунта / Выход.

### 4. Авторизация (JWT-подобные токены)
- `POST /auth/register`, `POST /auth/login` → `AuthResponse{Token, User}`.
- Токен — 32 случайных байта из `crypto/rand`, hex. Хранится в `TokenStore` (map + RWMutex).
- Пароль хешируется **bcrypt**.
- `AuthMiddleware` проверяет `Authorization: Bearer <token>`.
- `POST /auth/logout` удаляет токен.

### 5. Несколько клиентов
Запустите 2+ клиента в разных терминалах — данные общие (одна БД), сессии разные (свои токены). `net/http` создаёт горутину на соединение.

### 6. Тестирование
См. раздел ниже.

## Установка

```bash
cd lab9
go mod tidy
```

Если `go get modernc.org/sqlite` висит:
```bash
go env -w GOPROXY=https://goproxy.io,direct
go mod tidy
```

## Запуск

### Терминал 1 — сервер
```bash
cd lab9/server
go run . -storage=memory
# или: go run . -storage=sqlite -db=lab9.db
```

### Терминал 2, 3, … — клиенты
```bash
cd lab9/client
go run .
# или с другим сервером: go run . -server=http://localhost:9090 (или 8080)
```

## Тестирование

### Сценарий 1: регистрация → CRUD
```
1. Клиент 1 → 1 (Регистрация)
   Alice / alice@example.com / 30 / secret123
   ✅ Вы зарегистрированы и вошли как Alice

2. Клиент 1 → 1 (Список) → видит Alice

3. Клиент 1 → 3 (Создать) → Bob / bob@example.com / 25
   ✅ Создан

4. Клиент 1 → 1 (Список) → Alice, Bob
```

### Сценарий 2: несколько клиентов
```
1. Клиент 1 → регистрация Alice
2. Клиент 2 → регистрация Bob
3. Клиент 1 → 1 (Список) → видит Alice и Bob
4. Клиент 2 → 1 (Список) → видит Bob и Alice
5. Клиент 1 → 5 (Удалить) → 2 (Bob) → y
6. Клиент 2 → 1 (Список) → Bob исчез
```

### Сценарий 3: ошибки
| Ввод | Ожидание |
|------|----------|
| Пароль `123` при регистрации | `❌ validation failed: password must be at least 6 characters` |
| Регистрация с тем же email | `❌ HTTP 409: email already exists` |
| Вход с неверным паролем | `❌ HTTP 401: invalid credentials` |
| Создание с `age=-5` | `❌ validation failed: age must be >= 0` |
| `GET /users/999` | `❌ HTTP 404: user not found` |
| Сервер не запущен | `❌ сетевая ошибка: ...connection refused` |

### Сценарий 4: проверка защиты без токена
```powershell
curl.exe http://localhost:8080/users
# → 401 {"error":"unauthorized: no token"}

# С токеном
$r = Invoke-RestMethod -Uri http://localhost:8080/auth/login -Method Post `
  -ContentType "application/json" `
  -Body (@{email="alice@example.com"; password="secret123"} | ConvertTo-Json)
Invoke-RestMethod -Uri http://localhost:8080/users `
  -Headers @{ Authorization = "Bearer $($r.token)" }
```

## Частые проблемы

| Симптом | Решение |
|---------|---------|
| `401 no token` на `/users` | Сначала «Вход»/«Регистрация» |
| `401 invalid token` после рестарта сервера | `TokenStore` в памяти — перелогиньтесь |
| `no such column: password_hash` | Старая БД из lab8 — используйте `-db=lab9.db` |
| `connection refused` | Сервер не запущен |


package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "strconv"
    "strings"
    "time"
)

type User struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

type AuthResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}

type APIError struct {
    Code    int      `json:"-"`
    Message string   `json:"error"`
    Details []string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
    if len(e.Details) > 0 {
        return fmt.Sprintf("%s: %s", e.Message, strings.Join(e.Details, "; "))
    }
    return e.Message
}

type Client struct {
    baseURL string
    token   string
    http    *http.Client
    in      *bufio.Reader
}

func NewClient(baseURL string) *Client {
    return &Client{
        baseURL: strings.TrimRight(baseURL, "/"),
        http:    &http.Client{Timeout: 10 * time.Second},
        in:      bufio.NewReader(os.Stdin),
    }
}

func (c *Client) doRequest(method, path string, body any) ([]byte, int, error) {
    var reader io.Reader
    if body != nil {
        data, err := json.Marshal(body)
        if err != nil {
            return nil, 0, err
        }
        reader = bytes.NewReader(data)
    }

    req, err := http.NewRequest(method, c.baseURL+path, reader)
    if err != nil {
        return nil, 0, err
    }
    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }
    if c.token != "" {
        req.Header.Set("Authorization", "Bearer "+c.token)
    }

    resp, err := c.http.Do(req)
    if err != nil {
        return nil, 0, fmt.Errorf("сетевая ошибка: %w", err)
    }
    defer resp.Body.Close()

    data, _ := io.ReadAll(resp.Body)
    return data, resp.StatusCode, nil
}

func parseAPIError(data []byte, status int) error {
    var e APIError
    e.Code = status
    if err := json.Unmarshal(data, &e); err != nil || e.Message == "" {
        return fmt.Errorf("HTTP %d: %s", status, strings.TrimSpace(string(data)))
    }
    return &e
}

func (c *Client) register() {
    fmt.Println("\n--- Регистрация ---")
    name := c.readLine("Имя: ")
    email := c.readLine("Email: ")
    ageStr := c.readLine("Возраст: ")
    password := c.readLine("Пароль (мин. 6 символов): ")

    age, err := strconv.Atoi(ageStr)
    if err != nil {
        fmt.Println("❌ Возраст должен быть числом")
        return
    }

    body := map[string]any{"name": name, "email": email, "age": age, "password": password}
    data, status, err := c.doRequest("POST", "/auth/register", body)
    if err != nil {
        fmt.Println("❌", err)
        return
    }
    if status != http.StatusCreated {
        fmt.Println("❌", parseAPIError(data, status))
        return
    }

    var resp AuthResponse
    if err := json.Unmarshal(data, &resp); err != nil {
        fmt.Println("❌ Ошибка разбора ответа:", err)
        return
    }
    c.token = resp.Token
    fmt.Println("✅ Вы зарегистрированы и вошли как", resp.User.Name)
}

func (c *Client) login() {
    fmt.Println("\n--- Вход ---")
    email := c.readLine("Email: ")
    password := c.readLine("Пароль: ")

    body := map[string]any{"email": email, "password": password}
    data, status, err := c.doRequest("POST", "/auth/login", body)
    if err != nil {
        fmt.Println("❌", err)
        return
    }
    if status != http.StatusOK {
        fmt.Println("❌", parseAPIError(data, status))
        return
    }

    var resp AuthResponse
    json.Unmarshal(data, &resp)
    c.token = resp.Token
    fmt.Printf("✅ Добро пожаловать, %s!\n", resp.User.Name)
}

func (c *Client) logout() {
    if c.token == "" {
        fmt.Println("ℹ️  Вы не авторизованы")
        return
    }
    _, _, _ = c.doRequest("POST", "/auth/logout", nil)
    c.token = ""
    fmt.Println("👋 Вы вышли из системы")
}

func (c *Client) listUsers() {
    fmt.Println("\n--- Список пользователей ---")
    name := c.readLine("Фильтр по имени (Enter — без фильтра): ")
    pageStr := c.readLine("Страница (Enter — 1): ")
    limitStr := c.readLine("Лимит (Enter — 20): ")

    q := url.Values{}
    if name != "" {
        q.Set("name", name)
    }
    if pageStr != "" {
        q.Set("page", pageStr)
    }
    if limitStr != "" {
        q.Set("limit", limitStr)
    }
    path := "/users"
    if len(q) > 0 {
        path += "?" + q.Encode()
    }

    data, status, err := c.doRequest("GET", path, nil)
    if err != nil {
        fmt.Println("❌", err)
        return
    }
    if status != http.StatusOK {
        fmt.Println("❌", parseAPIError(data, status))
        return
    }

    var resp struct {
        Data  []User `json:"data"`
        Total int    `json:"total"`
        Page  int    `json:"page"`
        Limit int    `json:"limit"`
    }
    if err := json.Unmarshal(data, &resp); err != nil {
        fmt.Println("❌ Ошибка разбора:", err)
        return
    }
    printUsersTable(resp.Data)
    fmt.Printf("Всего: %d, страница %d, лимит %d\n", resp.Total, resp.Page, resp.Limit)
}

func (c *Client) getUser() {
    idStr := c.readLine("\nID пользователя: ")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        fmt.Println("❌ Неверный ID")
        return
    }

    data, status, err := c.doRequest("GET", "/users/"+idStr, nil)
    if err != nil {
        fmt.Println("❌", err)
        return
    }
    if status != http.StatusOK {
        fmt.Println("❌", parseAPIError(data, status))
        return
    }

    var u User
    json.Unmarshal(data, &u)
    printUser(u)
}

func (c *Client) createUser() {
    fmt.Println("\n--- Создать пользователя ---")
    name := c.readLine("Имя: ")
    email := c.readLine("Email: ")
    ageStr := c.readLine("Возраст: ")

    age, err := strconv.Atoi(ageStr)
    if err != nil {
        fmt.Println("❌ Возраст должен быть числом")
        return
    }

    body := map[string]any{"name": name, "email": email, "age": age}
    data, status, err := c.doRequest("POST", "/users", body)
    if err != nil {
        fmt.Println("❌", err)
        return
    }
    if status != http.StatusCreated {
        fmt.Println("❌", parseAPIError(data, status))
        return
    }

    var u User
    json.Unmarshal(data, &u)
    fmt.Println("✅ Создан:")
    printUser(u)
}

func (c *Client) updateUser() {
    idStr := c.readLine("\nID пользователя для обновления: ")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        fmt.Println("❌ Неверный ID")
        return
    }

    fmt.Println("Оставьте поле пустым, чтобы не менять:")
    name := c.readLine("Новое имя: ")
    email := c.readLine("Новый email: ")
    ageStr := c.readLine("Новый возраст: ")

    body := map[string]any{}
    if name != "" {
        body["name"] = name
    }
    if email != "" {
        body["email"] = email
    }
    if ageStr != "" {
        age, err := strconv.Atoi(ageStr)
        if err != nil {
            fmt.Println("❌ Возраст должен быть числом")
            return
        }
        body["age"] = age
    }
    if len(body) == 0 {
        fmt.Println("ℹ️  Ничего не изменено")
        return
    }

    data, status, err := c.doRequest("PUT", "/users/"+idStr, body)
    if err != nil {
        fmt.Println("❌", err)
        return
    }
    if status != http.StatusOK {
        fmt.Println("❌", parseAPIError(data, status))
        return
    }

    var u User
    json.Unmarshal(data, &u)
    fmt.Println("✅ Обновлён:")
    printUser(u)
}

func (c *Client) deleteUser() {
    idStr := c.readLine("\nID пользователя для удаления: ")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil || id <= 0 {
        fmt.Println("❌ Неверный ID")
        return
    }

    confirm := c.readLine(fmt.Sprintf("Удалить пользователя #%d? (y/N): ", id))
    if strings.ToLower(confirm) != "y" {
        fmt.Println("Отменено")
        return
    }

    data, status, err := c.doRequest("DELETE", "/users/"+idStr, nil)
    if err != nil {
        fmt.Println("❌", err)
        return
    }
    if status != http.StatusNoContent {
        fmt.Println("❌", parseAPIError(data, status))
        return
    }
    fmt.Println("✅ Удалён")
}

func printUsersTable(users []User) {
    if len(users) == 0 {
        fmt.Println("(пусто)")
        return
    }
    fmt.Printf("%-5s %-20s %-30s %-8s\n", "ID", "ИМЯ", "EMAIL", "ВОЗРАСТ")
    fmt.Println(strings.Repeat("─", 66))
    for _, u := range users {
        fmt.Printf("%-5d %-20s %-30s %-8d\n", u.ID, u.Name, u.Email, u.Age)
    }
}

func printUser(u User) {
    fmt.Printf("  ID:      %d\n", u.ID)
    fmt.Printf("  Имя:     %s\n", u.Name)
    fmt.Printf("  Email:   %s\n", u.Email)
    fmt.Printf("  Возраст: %d\n", u.Age)
}

func (c *Client) run() {
    for {
        fmt.Println()
        if c.token == "" {
            fmt.Println("=== Users Client (гость) ===")
            fmt.Println("1. Регистрация")
            fmt.Println("2. Вход")
            fmt.Println("0. Выход")
            fmt.Print("Выбор: ")
        } else {
            fmt.Println("=== Users Client (авторизован) ===")
            fmt.Println("1. Список пользователей")
            fmt.Println("2. Пользователь по ID")
            fmt.Println("3. Создать пользователя")
            fmt.Println("4. Обновить пользователя")
            fmt.Println("5. Удалить пользователя")
            fmt.Println("6. Выйти из аккаунта")
            fmt.Println("0. Выход из программы")
            fmt.Print("Выбор: ")
        }

        choice := strings.TrimSpace(c.mustReadLine())
        fmt.Println()

        if c.token == "" {
            switch choice {
            case "1":
                c.register()
            case "2":
                c.login()
            case "0":
                fmt.Println("Пока!")
                return
            default:
                fmt.Println("❌ Неверный выбор")
            }
        } else {
            switch choice {
            case "1":
                c.listUsers()
            case "2":
                c.getUser()
            case "3":
                c.createUser()
            case "4":
                c.updateUser()
            case "5":
                c.deleteUser()
            case "6":
                c.logout()
            case "0":
                fmt.Println("Пока!")
                return
            default:
                fmt.Println("❌ Неверный выбор")
            }
        }
    }
}

func (c *Client) readLine(prompt string) string {
    fmt.Print(prompt)
    line, _ := c.in.ReadString('\n')
    return strings.TrimRight(line, "\r\n")
}

func (c *Client) mustReadLine() string {
    line, _ := c.in.ReadString('\n')
    return strings.TrimRight(line, "\r\n")
}

func main() {
    baseURL := flag.String("server", "http://localhost:8080", "server URL")
    flag.Parse()

    c := NewClient(*baseURL)
    fmt.Println("Клиент запущен. Сервер:", c.baseURL)
    c.run()
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
    return &MemoryStorage{nextID: 1, users: make(map[int64]User)}
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

func (m *MemoryStorage) GetByEmail(_ context.Context, email string) (User, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    for _, u := range m.users {
        if u.Email == email {
            return u, nil
        }
    }
    return User{}, ErrNotFound
}

func (m *MemoryStorage) Create(_ context.Context, u User) (User, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
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
    if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
        return nil, fmt.Errorf("pragma: %w", err)
    }

    schema := `
    CREATE TABLE IF NOT EXISTS users (
        id            INTEGER PRIMARY KEY AUTOINCREMENT,
        name          TEXT    NOT NULL,
        email         TEXT    NOT NULL UNIQUE,
        age           INTEGER NOT NULL,
        password_hash TEXT    NOT NULL DEFAULT ''
    );`
    if _, err := db.Exec(schema); err != nil {
        return nil, fmt.Errorf("schema: %w", err)
    }

    return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) List(ctx context.Context, f Filter) ([]User, int, error) {
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

    var total int
    if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users"+where, args...).Scan(&total); err != nil {
        return nil, 0, err
    }

    query := "SELECT id, name, email, age, COALESCE(password_hash,'') FROM users" + where + " ORDER BY id"
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
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Age, &u.PasswordHash); err != nil {
            return nil, 0, err
        }
        users = append(users, u)
    }
    return users, total, rows.Err()
}

func (s *SQLiteStorage) Get(ctx context.Context, id int64) (User, error) {
    var u User
    err := s.db.QueryRowContext(ctx,
        "SELECT id, name, email, age, COALESCE(password_hash,'') FROM users WHERE id = ?", id).
        Scan(&u.ID, &u.Name, &u.Email, &u.Age, &u.PasswordHash)
    if errors.Is(err, sql.ErrNoRows) {
        return User{}, ErrNotFound
    }
    return u, err
}

func (s *SQLiteStorage) GetByEmail(ctx context.Context, email string) (User, error) {
    var u User
    err := s.db.QueryRowContext(ctx,
        "SELECT id, name, email, age, COALESCE(password_hash,'') FROM users WHERE email = ?", email).
        Scan(&u.ID, &u.Name, &u.Email, &u.Age, &u.PasswordHash)
    if errors.Is(err, sql.ErrNoRows) {
        return User{}, ErrNotFound
    }
    return u, err
}

func (s *SQLiteStorage) Create(ctx context.Context, u User) (User, error) {
    res, err := s.db.ExecContext(ctx,
        "INSERT INTO users(name, email, age, password_hash) VALUES(?, ?, ?, ?)",
        u.Name, u.Email, u.Age, u.PasswordHash)
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
        "UPDATE users SET name=?, email=?, age=?, password_hash=? WHERE id=?",
        u.Name, u.Email, u.Age, u.PasswordHash, id)
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

type User struct {
    ID           int64
    Name         string
    Email        string
    Age          int
    PasswordHash string
}

type Filter struct {
    NameLike string
    MinAge   int
    MaxAge   int
    Limit    int
    Offset   int
}

var (
    ErrNotFound       = errors.New("user not found")
    ErrDuplicateEmail = errors.New("email already exists")
)

type UserStorage interface {
    List(ctx context.Context, f Filter) ([]User, int, error)
    Get(ctx context.Context, id int64) (User, error)
    GetByEmail(ctx context.Context, email string) (User, error)
    Create(ctx context.Context, u User) (User, error)
    Update(ctx context.Context, id int64, u User) (User, error)
    Delete(ctx context.Context, id int64) error
    Close() error
}


package main

import (
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "errors"
    "net/http"
    "sync"

    "golang.org/x/crypto/bcrypt"
    "lab9/server/storage"
)

type RegisterRequest struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Age      int    `json:"age"`
    Password string `json:"password"`
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type AuthResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}

type TokenStore struct {
    mu     sync.RWMutex
    tokens map[string]int64
}

func NewTokenStore() *TokenStore {
    return &TokenStore{tokens: make(map[string]int64)}
}

func (t *TokenStore) Create(userID int64) string {
    b := make([]byte, 32)
    rand.Read(b)
    token := hex.EncodeToString(b)
    t.mu.Lock()
    t.tokens[token] = userID
    t.mu.Unlock()
    return token
}

func (t *TokenStore) Get(token string) (int64, bool) {
    t.mu.RLock()
    defer t.mu.RUnlock()
    id, ok := t.tokens[token]
    return id, ok
}

func (t *TokenStore) Delete(token string) {
    t.mu.Lock()
    delete(t.tokens, token)
    t.mu.Unlock()
}

type AuthHandler struct {
    storage storage.UserStorage
    tokens  *TokenStore
}

func NewAuthHandler(s storage.UserStorage, t *TokenStore) *AuthHandler {
    return &AuthHandler{storage: s, tokens: t}
}

// POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) error {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return NewBadRequest("invalid JSON: "+err.Error(), nil)
    }
    if len(req.Password) < 6 {
        return NewBadRequest("validation failed",
            []string{"password must be at least 6 characters"})
    }
    if err := validateCreate(CreateUserRequest{
        Name: req.Name, Email: req.Email, Age: req.Age,
    }); err != nil {
        return err
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    u, err := h.storage.Create(r.Context(), storage.User{
        Name: req.Name, Email: req.Email, Age: req.Age,
        PasswordHash: string(hash),
    })
    if errors.Is(err, storage.ErrDuplicateEmail) {
        return NewConflict("email already exists")
    }
    if err != nil {
        return err
    }

    token := h.tokens.Create(u.ID)
    writeJSON(w, http.StatusCreated, AuthResponse{Token: token, User: toAPIUser(u)})
    return nil
}

// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) error {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return NewBadRequest("invalid JSON: "+err.Error(), nil)
    }

    u, err := h.storage.GetByEmail(r.Context(), req.Email)
    if errors.Is(err, storage.ErrNotFound) {
        return &AppError{Code: http.StatusUnauthorized, Message: "invalid credentials"}
    }
    if err != nil {
        return err
    }

    if err := bcrypt.CompareHashAndPassword(
        []byte(u.PasswordHash), []byte(req.Password),
    ); err != nil {
        return &AppError{Code: http.StatusUnauthorized, Message: "invalid credentials"}
    }

    token := h.tokens.Create(u.ID)
    writeJSON(w, http.StatusOK, AuthResponse{Token: token, User: toAPIUser(u)})
    return nil
}

// POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) error {
    if token := extractBearer(r); token != "" {
        h.tokens.Delete(token)
    }
    w.WriteHeader(http.StatusNoContent)
    return nil
}

func extractBearer(r *http.Request) string {
    h := r.Header.Get("Authorization")
    if len(h) < 7 || h[:7] != "Bearer " {
        return ""
    }
    return h[7:]
}

func AuthMiddleware(tokens *TokenStore) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractBearer(r)
            if token == "" {
                writeError(w, &AppError{Code: http.StatusUnauthorized, Message: "unauthorized: no token"})
                return
            }
            if _, ok := tokens.Get(token); !ok {
                writeError(w, &AppError{Code: http.StatusUnauthorized, Message: "unauthorized: invalid token"})
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}


package main

import (
    "encoding/json"
    "errors"
    "net/http"
)

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

func writeError(w http.ResponseWriter, err error) {
    var appErr *AppError
    if !errors.As(err, &appErr) {
        appErr = &AppError{Code: http.StatusInternalServerError, Message: "internal server error"}
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(appErr.Code)
    json.NewEncoder(w).Encode(appErr)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}


package main

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"

    "lab9/server/storage"
)

type APIHandler func(w http.ResponseWriter, r *http.Request) error

func (h APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := h(w, r); err != nil {
        writeError(w, err)
    }
}

func toAPIUser(u storage.User) User {
    return User{ID: u.ID, Name: u.Name, Email: u.Email, Age: u.Age}
}

type UserHandler struct {
    storage storage.UserStorage
}

func NewUserHandler(s storage.UserStorage) *UserHandler {
    return &UserHandler{storage: s}
}

// GET /users
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) error {
    f, err := parseFilter(r)
    if err != nil {
        return err
    }
    users, total, err := h.storage.List(r.Context(), f)
    if err != nil {
        return err
    }

    apiUsers := make([]User, len(users))
    for i, u := range users {
        apiUsers[i] = toAPIUser(u)
    }

    page := f.Offset/f.Limit + 1
    writeJSON(w, http.StatusOK, map[string]any{
        "data":  apiUsers,
        "total": total,
        "page":  page,
        "limit": f.Limit,
    })
    return nil
}

// GET /users/{id}
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
    writeJSON(w, http.StatusOK, toAPIUser(u))
    return nil
}

// POST /users
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
    writeJSON(w, http.StatusCreated, toAPIUser(u))
    return nil
}

// PUT /users/{id}
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

    currentAPI := toAPIUser(current)
    updated, err := validateUpdate(req, currentAPI)
    if err != nil {
        return err
    }

    result, err := h.storage.Update(r.Context(), id, storage.User{
        Name: updated.Name, Email: updated.Email, Age: updated.Age,
        PasswordHash: current.PasswordHash,
    })
    if errors.Is(err, storage.ErrDuplicateEmail) {
        return NewConflict("email already exists")
    }
    if err != nil {
        return err
    }
    writeJSON(w, http.StatusOK, toAPIUser(result))
    return nil
}

// DELETE /users/{id}
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

func parseID(r *http.Request) (int64, error) {
    s := r.PathValue("id")
    id, err := strconv.ParseInt(s, 10, 64)
    if err != nil || id <= 0 {
        return 0, NewBadRequest("invalid id", nil)
    }
    return id, nil
}

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
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "lab9/server/storage"
)

func main() {
    var (
        addr        = flag.String("addr", ":8080", "server address")
        storageKind = flag.String("storage", "sqlite", "storage: memory | sqlite")
        dbPath      = flag.String("db", "lab9.db", "path to sqlite db")
    )
    flag.Parse()

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

    tokens := NewTokenStore()
    authHandler := NewAuthHandler(store, tokens)
    userHandler := NewUserHandler(store)

    authMW := AuthMiddleware(tokens)

    mux := http.NewServeMux()

    // публичные
    mux.Handle("POST /auth/register", APIHandler(authHandler.Register))
    mux.Handle("POST /auth/login", APIHandler(authHandler.Login))

    // защищённые
    mux.Handle("POST /auth/logout", authMW(APIHandler(authHandler.Logout)))
    mux.Handle("GET /users", authMW(APIHandler(userHandler.List)))
    mux.Handle("POST /users", authMW(APIHandler(userHandler.Create)))
    mux.Handle("GET /users/{id}", authMW(APIHandler(userHandler.Get)))
    mux.Handle("PUT /users/{id}", authMW(APIHandler(userHandler.Update)))
    mux.Handle("DELETE /users/{id}", authMW(APIHandler(userHandler.Delete)))

    handler := loggingMiddleware(recoveryMiddleware(mux))

    srv := &http.Server{
        Addr:         *addr,
        Handler:      handler,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    go func() {
        log.Printf("Listening on %s (storage=%s)", *addr, *storageKind)
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
        log.Println("shutdown:", err)
    }
    log.Println("Server stopped")
}


package main

import (
    "log"
    "net/http"
    "time"
)

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        rw := &statusRecorder{ResponseWriter: w, status: 200}
        next.ServeHTTP(rw, r)
        log.Printf("%s %s -> %d (%v)", r.Method, r.URL.RequestURI(), rw.status, time.Since(start))
    })
}

func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if v := recover(); v != nil {
                log.Printf("PANIC: %v", v)
                writeError(w, &AppError{Code: http.StatusInternalServerError, Message: "internal server error"})
            }
        }()
        next.ServeHTTP(w, r)
    })
}

type statusRecorder struct {
    http.ResponseWriter
    status int
}

func (s *statusRecorder) WriteHeader(code int) {
    s.status = code
    s.ResponseWriter.WriteHeader(code)
}


package main

// User — модель API (без пароля).
type User struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

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

func validateCreate(req CreateUserRequest) error {
    errs := collectErrors(req.Name, req.Email, req.Age)
    if len(errs) > 0 {
        return NewBadRequest("validation failed", errs)
    }
    return nil
}

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
        errs = append(errs, fmt.Sprintf("name is too long (max 100, got %d)", len([]rune(name))))
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