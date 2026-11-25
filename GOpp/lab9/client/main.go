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