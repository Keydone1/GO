package main

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"

    "lab8/storage"
)

// APIHandler — handler, возвращающий ошибку.
// ServeHTTP сам вызывает writeError при ошибке → централизованно.
type APIHandler func(w http.ResponseWriter, r *http.Request) error

func (h APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := h(w, r); err != nil {
        writeError(w, err)
    }
}

type UserHandler struct {
    storage storage.UserStorage
}

func NewUserHandler(s storage.UserStorage) *UserHandler {
    return &UserHandler{storage: s}
}

// ---------- GET /users ----------
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) error {
    f, err := parseFilter(r)
    if err != nil {
        return err
    }
    users, total, err := h.storage.List(r.Context(), f)
    if err != nil {
        return err
    }

    page := f.Offset/f.Limit + 1
    writeJSON(w, http.StatusOK, map[string]any{
        "data":  users,
        "total": total,
        "page":  page,
        "limit": f.Limit,
    })
    return nil
}

// ---------- GET /users/{id} ----------
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) error {
    id, err := parseID(r)
    if err != nil {
        return err
    }
    u, err := h.storage.Get(r.Context(), id)
    if errors.Is(err, storage.ErrNotFound) {
        return NewNotFound("user not found")
    }
    if err != nil {
        return err
    }
    writeJSON(w, http.StatusOK, u)
    return nil
}

// ---------- POST /users ----------
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) error {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return NewBadRequest("invalid JSON: "+err.Error(), nil)
    }
    if err := validateCreate(req); err != nil {
        return err
    }

    u, err := h.storage.Create(r.Context(), storage.User{
        Name: req.Name, Email: req.Email, Age: req.Age,
    })
    if errors.Is(err, storage.ErrDuplicateEmail) {
        return NewConflict("email already exists")
    }
    if err != nil {
        return err
    }
    writeJSON(w, http.StatusCreated, u)
    return nil
}

// ---------- PUT /users/{id} ----------
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) error {
    id, err := parseID(r)
    if err != nil {
        return err
    }

    current, err := h.storage.Get(r.Context(), id)
    if errors.Is(err, storage.ErrNotFound) {
        return NewNotFound("user not found")
    }
    if err != nil {
        return err
    }

    var req UpdateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return NewBadRequest("invalid JSON: "+err.Error(), nil)
    }

    // Конвертация storage.User → main.User для валидации
    currentAPI := User{
        ID:    current.ID,
        Name:  current.Name,
        Email: current.Email,
        Age:   current.Age,
    }

    // ← переменная называется updated (НЕ result!)
    updated, err := validateUpdate(req, currentAPI)
    if err != nil {
        return err
    }

    // ← переменная называется result
    result, err := h.storage.Update(r.Context(), id, storage.User{
        Name:  updated.Name,
        Email: updated.Email,
        Age:   updated.Age,
    })
    if errors.Is(err, storage.ErrDuplicateEmail) {
        return NewConflict("email already exists")
    }
    if err != nil {
        return err
    }

    writeJSON(w, http.StatusOK, result)
    return nil
}

// ---------- DELETE /users/{id} ----------
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) error {
    id, err := parseID(r)
    if err != nil {
        return err
    }
    if err := h.storage.Delete(r.Context(), id); err != nil {
        if errors.Is(err, storage.ErrNotFound) {
            return NewNotFound("user not found")
        }
        return err
    }
    w.WriteHeader(http.StatusNoContent)
    return nil
}

// ---------- helpers ----------

func parseID(r *http.Request) (int64, error) {
    s := r.PathValue("id")
    id, err := strconv.ParseInt(s, 10, 64)
    if err != nil || id <= 0 {
        return 0, NewBadRequest("invalid id", nil)
    }
    return id, nil
}

// parseFilter — пагинация и фильтрация из query-параметров.
func parseFilter(r *http.Request) (storage.Filter, error) {
    f := storage.Filter{Limit: 20, Offset: 0}
    q := r.URL.Query()

    if v := q.Get("name"); v != "" {
        f.NameLike = v
    }
    if v := q.Get("min_age"); v != "" {
        n, err := strconv.Atoi(v)
        if err != nil || n < 0 {
            return f, NewBadRequest("invalid min_age", nil)
        }
        f.MinAge = n
    }
    if v := q.Get("max_age"); v != "" {
        n, err := strconv.Atoi(v)
        if err != nil || n < 0 {
            return f, NewBadRequest("invalid max_age", nil)
        }
        f.MaxAge = n
    }

    page := 1
    if v := q.Get("page"); v != "" {
        n, err := strconv.Atoi(v)
        if err != nil || n < 1 {
            return f, NewBadRequest("invalid page", nil)
        }
        page = n
    }

    if v := q.Get("limit"); v != "" {
        n, err := strconv.Atoi(v)
        if err != nil || n < 1 || n > 100 {
            return f, NewBadRequest("invalid limit (1..100)", nil)
        }
        f.Limit = n
    }

    f.Offset = (page - 1) * f.Limit
    return f, nil
}