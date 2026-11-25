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