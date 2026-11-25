package storage

import (
    "context"
    "errors"
)

// User — модель на уровне хранилища (может совпадать с моделью API).
type User struct {
    ID    int64
    Name  string
    Email string
    Age   int
}

// Filter — параметры фильтрации и пагинации.
type Filter struct {
    NameLike string // LIKE %NameLike%
    MinAge   int    // 0 = не фильтровать
    MaxAge   int    // 0 = не фильтровать
    Limit    int    // 0 = не ограничивать
    Offset   int
}

// Ошибки хранилища — не HTTP, а «бизнес-ошибки».
// Handlers преобразуют их в HTTP-коды.
var (
    ErrNotFound       = errors.New("user not found")
    ErrDuplicateEmail = errors.New("email already exists")
)

// UserStorage — абстракция над БД.
// Позволяет переключать Memory <-> SQLite без изменения handlers.
type UserStorage interface {
    List(ctx context.Context, f Filter) (users []User, total int, err error)
    Get(ctx context.Context, id int64) (User, error)
    Create(ctx context.Context, u User) (User, error)
    Update(ctx context.Context, id int64, u User) (User, error)
    Delete(ctx context.Context, id int64) error
    Close() error
}