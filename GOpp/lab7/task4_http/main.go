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