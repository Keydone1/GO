package main

import (
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "errors"
    "net/http"
    "sync"

    "golang.org/x/crypto/bcrypt"
    "lab9/server/storage"
)

type RegisterRequest struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Age      int    `json:"age"`
    Password string `json:"password"`
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type AuthResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}

type TokenStore struct {
    mu     sync.RWMutex
    tokens map[string]int64
}

func NewTokenStore() *TokenStore {
    return &TokenStore{tokens: make(map[string]int64)}
}

func (t *TokenStore) Create(userID int64) string {
    b := make([]byte, 32)
    rand.Read(b)
    token := hex.EncodeToString(b)
    t.mu.Lock()
    t.tokens[token] = userID
    t.mu.Unlock()
    return token
}

func (t *TokenStore) Get(token string) (int64, bool) {
    t.mu.RLock()
    defer t.mu.RUnlock()
    id, ok := t.tokens[token]
    return id, ok
}

func (t *TokenStore) Delete(token string) {
    t.mu.Lock()
    delete(t.tokens, token)
    t.mu.Unlock()
}

type AuthHandler struct {
    storage storage.UserStorage
    tokens  *TokenStore
}

func NewAuthHandler(s storage.UserStorage, t *TokenStore) *AuthHandler {
    return &AuthHandler{storage: s, tokens: t}
}

// POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) error {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return NewBadRequest("invalid JSON: "+err.Error(), nil)
    }
    if len(req.Password) < 6 {
        return NewBadRequest("validation failed",
            []string{"password must be at least 6 characters"})
    }
    if err := validateCreate(CreateUserRequest{
        Name: req.Name, Email: req.Email, Age: req.Age,
    }); err != nil {
        return err
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    u, err := h.storage.Create(r.Context(), storage.User{
        Name: req.Name, Email: req.Email, Age: req.Age,
        PasswordHash: string(hash),
    })
    if errors.Is(err, storage.ErrDuplicateEmail) {
        return NewConflict("email already exists")
    }
    if err != nil {
        return err
    }

    token := h.tokens.Create(u.ID)
    writeJSON(w, http.StatusCreated, AuthResponse{Token: token, User: toAPIUser(u)})
    return nil
}

// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) error {
    var req LoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        return NewBadRequest("invalid JSON: "+err.Error(), nil)
    }

    u, err := h.storage.GetByEmail(r.Context(), req.Email)
    if errors.Is(err, storage.ErrNotFound) {
        return &AppError{Code: http.StatusUnauthorized, Message: "invalid credentials"}
    }
    if err != nil {
        return err
    }

    if err := bcrypt.CompareHashAndPassword(
        []byte(u.PasswordHash), []byte(req.Password),
    ); err != nil {
        return &AppError{Code: http.StatusUnauthorized, Message: "invalid credentials"}
    }

    token := h.tokens.Create(u.ID)
    writeJSON(w, http.StatusOK, AuthResponse{Token: token, User: toAPIUser(u)})
    return nil
}

// POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) error {
    if token := extractBearer(r); token != "" {
        h.tokens.Delete(token)
    }
    w.WriteHeader(http.StatusNoContent)
    return nil
}

func extractBearer(r *http.Request) string {
    h := r.Header.Get("Authorization")
    if len(h) < 7 || h[:7] != "Bearer " {
        return ""
    }
    return h[7:]
}

func AuthMiddleware(tokens *TokenStore) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractBearer(r)
            if token == "" {
                writeError(w, &AppError{Code: http.StatusUnauthorized, Message: "unauthorized: no token"})
                return
            }
            if _, ok := tokens.Get(token); !ok {
                writeError(w, &AppError{Code: http.StatusUnauthorized, Message: "unauthorized: invalid token"})
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}