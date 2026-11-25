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