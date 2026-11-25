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