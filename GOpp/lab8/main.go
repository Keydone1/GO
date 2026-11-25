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