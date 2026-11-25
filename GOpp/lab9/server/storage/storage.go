package storage

import (
    "context"
    "errors"
)

type User struct {
    ID           int64
    Name         string
    Email        string
    Age          int
    PasswordHash string
}

type Filter struct {
    NameLike string
    MinAge   int
    MaxAge   int
    Limit    int
    Offset   int
}

var (
    ErrNotFound       = errors.New("user not found")
    ErrDuplicateEmail = errors.New("email already exists")
)

type UserStorage interface {
    List(ctx context.Context, f Filter) ([]User, int, error)
    Get(ctx context.Context, id int64) (User, error)
    GetByEmail(ctx context.Context, email string) (User, error)
    Create(ctx context.Context, u User) (User, error)
    Update(ctx context.Context, id int64, u User) (User, error)
    Delete(ctx context.Context, id int64) error
    Close() error
}