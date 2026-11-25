package storage

import (
    "context"
    "sort"
    "strings"
    "sync"
)

type MemoryStorage struct {
    mu     sync.RWMutex
    nextID int64
    users  map[int64]User
}

func NewMemoryStorage() *MemoryStorage {
    return &MemoryStorage{
        nextID: 1,
        users:  make(map[int64]User),
    }
}

func (m *MemoryStorage) List(_ context.Context, f Filter) ([]User, int, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    var filtered []User
    for _, u := range m.users {
        if f.NameLike != "" && !strings.Contains(strings.ToLower(u.Name), strings.ToLower(f.NameLike)) {
            continue
        }
        if f.MinAge > 0 && u.Age < f.MinAge {
            continue
        }
        if f.MaxAge > 0 && u.Age > f.MaxAge {
            continue
        }
        filtered = append(filtered, u)
    }

    sort.Slice(filtered, func(i, j int) bool { return filtered[i].ID < filtered[j].ID })

    total := len(filtered)

    // Пагинация
    if f.Offset >= len(filtered) {
        filtered = []User{}
    } else {
        filtered = filtered[f.Offset:]
        if f.Limit > 0 && f.Limit < len(filtered) {
            filtered = filtered[:f.Limit]
        }
    }

    return filtered, total, nil
}

func (m *MemoryStorage) Get(_ context.Context, id int64) (User, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    u, ok := m.users[id]
    if !ok {
        return User{}, ErrNotFound
    }
    return u, nil
}

func (m *MemoryStorage) Create(_ context.Context, u User) (User, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    // проверка уникальности email
    for _, existing := range m.users {
        if existing.Email == u.Email {
            return User{}, ErrDuplicateEmail
        }
    }
    u.ID = m.nextID
    m.nextID++
    m.users[u.ID] = u
    return u, nil
}

func (m *MemoryStorage) Update(_ context.Context, id int64, u User) (User, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    existing, ok := m.users[id]
    if !ok {
        return User{}, ErrNotFound
    }
    for otherID, other := range m.users {
        if otherID != id && other.Email == u.Email {
            return User{}, ErrDuplicateEmail
        }
    }
    u.ID = existing.ID
    m.users[id] = u
    return u, nil
}

func (m *MemoryStorage) Delete(_ context.Context, id int64) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    if _, ok := m.users[id]; !ok {
        return ErrNotFound
    }
    delete(m.users, id)
    return nil
}

func (m *MemoryStorage) Close() error { return nil }