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