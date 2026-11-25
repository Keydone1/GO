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