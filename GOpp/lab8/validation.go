package main

import (
    "fmt"
    "regexp"
    "strings"
)

var emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// validateCreate проверяет тело POST /users.
func validateCreate(req CreateUserRequest) error {
    errs := collectErrors(req.Name, req.Email, req.Age)
    if len(errs) > 0 {
        return NewBadRequest("validation failed", errs)
    }
    return nil
}

// validateUpdate проверяет тело PUT /users/{id}.
// Объединяем с текущим пользователем, чтобы получить итоговое состояние.
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
        errs = append(errs, fmt.Sprintf("name is too long (max 100 chars, got %d)", len([]rune(name))))
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