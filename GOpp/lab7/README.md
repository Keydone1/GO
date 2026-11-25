https://github.com/Keydone1/GO

## Структура
```
lab7/
─ go.mod
─ task1_tcp_server/main.go       :8081
─ task2_tcp_client/main.go       :8081
─ task3_tcp_async/main.go        :8082
─ task4_http/main.go             :8080
─ task5_http_routing/main.go     :8081 (поменять на :8083, если конфликт)
─ task6_websocket/main.go        :8082 (поменять на :8084, если конфликт)
```

## Задания

### 1. TCP-сервер
`net.Listen("tcp", ":8081")` → `Accept` в цикле → `go handleConn(conn)`. Читает строку через `bufio.Reader.ReadString('\n')`, отвечает `ACK: ...`.

### 2. TCP-клиент
`net.Dial("tcp", "localhost:8081")` → отправляет введённую строку → читает ответ → закрывает.

### 3. Асинхронный TCP + graceful shutdown
- `context.WithCancel` + `signal.Notify(SIGINT)` — корректная остановка.
- `sync.WaitGroup` + `time.After(5s)` — таймаут на завершение.
- `atomic.AddInt64` — счётчик активных соединений.

### 4. Базовый HTTP
`http.HandleFunc`:
- `GET /hello` → приветствие.
- `POST /data` → `json.Unmarshal` в структуру, вывод в консоль, ответ JSON.

### 5. HTTP + middleware
- `mux.HandleFunc("GET /hello", ...)` — Go 1.22+.
- `loggingMiddleware(next http.Handler)` — логирует метод, URL, время (`time.Since(start)`).

### 6. WebSocket-чат
`gorilla/websocket`, hub-паттерн: `register`/`unregister`/`broadcast` через `select`. На клиента — `readPump` + `writePump`.

## Установка (только для задачи 6)

```bash
cd lab7
go get github.com/gorilla/websocket
```

## Запуск

### Задачи 1 + 2
Терминал 1:
```bash
cd lab7/task1_tcp_server && go run .
```
Терминал 2:
```bash
cd lab7/task2_tcp_client && go run .
```

### Задача 3
```bash
cd lab7/task3_tcp_async && go run .
```
Тест: `nc localhost 8082` из двух терминалов, затем `Ctrl+C` на сервере — увидите graceful shutdown.

### Задача 4
```bash
cd lab7/task4_http && go run .
```
```bash
curl http://localhost:8080/hello
curl -X POST http://localhost:8080/data \
  -H "Content-Type: application/json" \
  -d '{"name":"Иван","value":42}'
```

### Задача 5
```bash
cd lab7/task5_http_routing && go run .
```
```bash
curl http://localhost:8081/hello
curl http://localhost:8081/time
curl -X POST http://localhost:8081/data -d '{"x":1}'
```

### Задача 6
```bash
cd lab7/task6_websocket && go run .
```
Откройте `http://localhost:8082` в **двух** окнах браузера — сообщения из одного видны в другом.

## Тестирование

| Задача | Ожидание |
|--------|----------|
| 1+2 | Клиент: `Ответ сервера: ACK: сообщение получено` |
| 3 | Ctrl+C → `Все соединения закрыты штатно` (или таймаут) |
| 4 | `GET /hello` → 200, `POST /data` → 200 + JSON |
| 5 | В логах сервера: `GET /hello -> 200 (45µs)` и т.д. |
| 6 | Два браузера обмениваются сообщениями |

## Частые проблемы

- **`address already in use`** — порт занят. Поменяйте `:8080` → `:9090` и т.д.
- **WebSocket 403** — проверьте `CheckOrigin: return true`.
- **`connection refused` у клиента** — сервер не запущен.


package main

import (
    "bufio"
    "fmt"
    "log"
    "net"
    "strings"
)

func main() {
    ln, err := net.Listen("tcp", ":8081")
    if err != nil {
        log.Fatal("Listen:", err)
    }
    defer ln.Close()
    fmt.Println("TCP-сервер слушает :8081")

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Println("Accept:", err)
            continue
        }
        // Каждое соединение — в своей горутине
        go handleConn(conn)
    }
}

func handleConn(conn net.Conn) {
    defer conn.Close()
    fmt.Printf("[+] Подключён %s\n", conn.RemoteAddr())

    reader := bufio.NewReader(conn)
    msg, err := reader.ReadString('\n')
    if err != nil {
        log.Println("Read:", err)
        return
    }

    fmt.Printf("Получено от %s: %q\n", conn.RemoteAddr(), strings.TrimSpace(msg))

    // Отправляем подтверждение
    fmt.Fprintf(conn, "ACK: сообщение получено (%d байт)\n", len(msg))
    fmt.Printf("[-] Отключён %s\n", conn.RemoteAddr())
}

package main

import (
    "bufio"
    "fmt"
    "log"
    "net"
    "os"
    "strings"
)

func main() {
    conn, err := net.Dial("tcp", "localhost:8081")
    if err != nil {
        log.Fatal("Dial:", err)
    }
    defer conn.Close()

    fmt.Print("Введите сообщение: ")
    reader := bufio.NewReader(os.Stdin)
    msg, _ := reader.ReadString('\n')
    msg = strings.TrimRight(msg, "\r\n")

    // Отправляем сообщение с переводом строки — сервер читает до '\n'
    fmt.Fprintf(conn, "%s\n", msg)

    // Ждём ответ
    resp, err := bufio.NewReader(conn).ReadString('\n')
    if err != nil {
        log.Fatal("Read:", err)
    }
    fmt.Println("Ответ сервера:", strings.TrimSpace(resp))
}

package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "net"
    "os"
    "os/signal"
    "strings"
    "sync"
    "sync/atomic"
    "syscall"
    "time"
)

func main() {
    ln, err := net.Listen("tcp", ":8082")
    if err != nil {
        log.Fatal("Listen:", err)
    }
    fmt.Println("Асинхронный TCP-сервер слушает :8082")
    fmt.Println("Для остановки нажмите Ctrl+C")

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    var wg sync.WaitGroup
    var activeConnections int64

    // ---- Accept loop ----
    go func() {
        for {
            conn, err := ln.Accept()
            if err != nil {
                select {
                case <-ctx.Done():
                    return // ожидаемое завершение после Close()
                default:
                    log.Println("Accept:", err)
                    continue
                }
            }
            atomic.AddInt64(&activeConnections, 1)
            wg.Add(1)
            go handleConn(ctx, conn, &wg, &activeConnections)
        }
    }()

    // ---- Ожидание сигнала ----
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh
    log.Println("Получен сигнал, начинаем graceful shutdown...")

    // Перестаём принимать новые соединения
    cancel()
    ln.Close()

    // Ждём завершения активных соединений с таймаутом
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        log.Println("Все соединения закрыты штатно")
    case <-time.After(5 * time.Second):
        log.Printf("Таймаут! Осталось активных соединений: %d\n",
            atomic.LoadInt64(&activeConnections))
    }
}

func handleConn(ctx context.Context, conn net.Conn, wg *sync.WaitGroup, active *int64) {
    defer wg.Done()
    defer conn.Close()
    defer atomic.AddInt64(active, -1)

    remote := conn.RemoteAddr().String()
    log.Printf("[+] %s подключён", remote)

    reader := bufio.NewReader(conn)
    for {
        // Проверяем отмену контекста — если shutdown, выходим
        select {
        case <-ctx.Done():
            fmt.Fprintf(conn, "Сервер выключается, до свидания!\n")
            return
        default:
        }

        // Таймаут на чтение, чтобы не блокироваться навсегда
        conn.SetReadDeadline(time.Now().Add(30 * time.Second))
        msg, err := reader.ReadString('\n')
        if err != nil {
            log.Printf("[-] %s отключён: %v", remote, err)
            return
        }

        msg = strings.TrimSpace(msg)
        if msg == "quit" {
            fmt.Fprintf(conn, "До свидания!\n")
            return
        }
        log.Printf("[%s] %s", remote, msg)
        fmt.Fprintf(conn, "ACK: %s (%d символов)\n", msg, len([]rune(msg)))
    }
}

package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
)

type DataRequest struct {
    Name  string `json:"name"`
    Value int    `json:"value"`
}

func main() {
    http.HandleFunc("/hello", helloHandler)
    http.HandleFunc("/data", dataHandler)

    fmt.Println("HTTP-сервер слушает :8080")
    fmt.Println("  GET  http://localhost:8080/hello")
    fmt.Println("  POST http://localhost:8080/data")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
        return
    }
    fmt.Fprintln(w, "Привет, мир! Это GET /hello")
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
        return
    }

    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Ошибка чтения тела", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    var data DataRequest
    if err := json.Unmarshal(body, &data); err != nil {
        http.Error(w, "Невалидный JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

    fmt.Printf("[POST /data] Получено: %+v\n", data)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "ok",
        "message": fmt.Sprintf("Получено: name=%s, value=%d", data.Name, data.Value),
    })
}

package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "time"
)

// loggingMiddleware логирует метод, URL и время выполнения каждого запроса.
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s — %v", r.Method, r.URL.Path, time.Since(start))
    })
}

func main() {
    mux := http.NewServeMux()

    // Go 1.22+: метод прямо в паттерне
    mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Привет из GET /hello!")
    })

    mux.HandleFunc("GET /time", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Сейчас: %s\n", time.Now().Format(time.RFC1123))
    })

    mux.HandleFunc("POST /data", func(w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)
        defer r.Body.Close()

        var data map[string]any
        if err := json.Unmarshal(body, &data); err != nil {
            http.Error(w, "Невалидный JSON", http.StatusBadRequest)
            return
        }
        log.Printf("[POST /data] %v", data)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]any{
            "status":  "ok",
            "received": data,
        })
    })

    // 404 для всех остальных путей
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.NotFound(w, r)
    })

    // Оборачиваем весь mux в middleware
    handler := loggingMiddleware(mux)

    addr := ":8081"
    fmt.Println("HTTP-сервер с middleware слушает", addr)
    fmt.Println("  GET  /hello")
    fmt.Println("  GET  /time")
    fmt.Println("  POST /data")
    log.Fatal(http.ListenAndServe(addr, handler))
}

package main

import (
    "log"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true }, // для локальной разработки
}

// Client — один WebSocket-клиент
type Client struct {
    conn *websocket.Conn
    send chan []byte
}

// Hub — центр, хранящий всех клиентов
type Hub struct {
    mu      sync.RWMutex
    clients map[*Client]bool
    broadcast chan []byte
    register  chan *Client
    unregister chan *Client
}

func newHub() *Hub {
    return &Hub{
        clients:    make(map[*Client]bool),
        broadcast:  make(chan []byte, 256),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
}

func (h *Hub) run() {
    for {
        select {
        case c := <-h.register:
            h.mu.Lock()
            h.clients[c] = true
            h.mu.Unlock()
            log.Printf("[+] Клиент подключён. Всего: %d", len(h.clients))

        case c := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[c]; ok {
                delete(h.clients, c)
                close(c.send)
            }
            h.mu.Unlock()
            log.Printf("[-] Клиент отключён. Всего: %d", len(h.clients))

        case msg := <-h.broadcast:
            h.mu.RLock()
            for c := range h.clients {
                select {
                case c.send <- msg:
                default:
                    // Клиент не успевает читать — отключаем
                    close(c.send)
                    delete(h.clients, c)
                }
            }
            h.mu.RUnlock()
        }
    }
}

// readPump читает сообщения от клиента и отправляет их в broadcast
func (c *Client) readPump(h *Hub) {
    defer func() {
        h.unregister <- c
        c.conn.Close()
    }()
    for {
        _, msg, err := c.conn.ReadMessage()
        if err != nil {
            log.Println("read:", err)
            return
        }
        log.Printf("Получено сообщение: %s", msg)
        h.broadcast <- msg
    }
}

// writePump отправляет сообщения клиенту
func (c *Client) writePump() {
    defer c.conn.Close()
    for msg := range c.send {
        if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
            log.Println("write:", err)
            return
        }
    }
}

func serveWs(h *Hub, w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("upgrade:", err)
        return
    }
    c := &Client{conn: conn, send: make(chan []byte, 256)}
    h.register <- c
    go c.writePump()
    c.readPump(h) // блокирующий вызов — читаем до отключения
}

func main() {
    h := newHub()
    go h.run()

    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        serveWs(h, w, r)
    })

    // Простая HTML-страница для теста
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Write([]byte(chatHTML))
    })

    log.Println("WebSocket-чат: http://localhost:8082")
    log.Fatal(http.ListenAndServe(":8082", nil))
}

const chatHTML = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Чат</title>
<style>
  body{font-family:monospace;max-width:600px;margin:20px auto;}
  #log{border:1px solid #ccc;height:300px;overflow:auto;padding:8px;margin-bottom:8px;}
  #msg{width:80%;padding:6px;}
</style></head><body>
<h2>WebSocket-чат</h2>
<div id="log"></div>
<input id="msg" placeholder="Введите сообщение..." autofocus>
<button onclick="send()">Отправить</button>
<script>
const ws = new WebSocket("ws://" + location.host + "/ws");
const log = document.getElementById("log");
ws.onmessage = e => {
  const p = document.createElement("div");
  p.textContent = "📨 " + e.data;
  log.appendChild(p);
  log.scrollTop = log.scrollHeight;
};
ws.onopen  = () => log.innerHTML += "<div>✅ Подключено</div>";
ws.onclose = () => log.innerHTML += "<div>❌ Отключено</div>";
function send() {
  const m = document.getElementById("msg").value;
  if (m) { ws.send(m); document.getElementById("msg").value = ""; }
}
document.getElementById("msg").addEventListener("keydown", e => {
  if (e.key === "Enter") send();
});
</script>
</body></html>`