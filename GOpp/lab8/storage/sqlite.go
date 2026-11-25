package storage

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "strings"

    _ "modernc.org/sqlite"
)

type SQLiteStorage struct {
    db *sql.DB
}

func NewSQLiteStorage(path string) (*SQLiteStorage, error) {
    db, err := sql.Open("sqlite", path)
    if err != nil {
        return nil, fmt.Errorf("open db: %w", err)
    }
    // SQLite: одно соединение на запись — включаем WAL для параллелизма
    if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
        return nil, fmt.Errorf("pragma: %w", err)
    }

    schema := `
    CREATE TABLE IF NOT EXISTS users (
        id    INTEGER PRIMARY KEY AUTOINCREMENT,
        name  TEXT    NOT NULL,
        email TEXT    NOT NULL UNIQUE,
        age   INTEGER NOT NULL
    );`
    if _, err := db.Exec(schema); err != nil {
        return nil, fmt.Errorf("schema: %w", err)
    }

    return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) List(ctx context.Context, f Filter) ([]User, int, error) {
    // Динамически строим WHERE и args
    var conds []string
    var args []any

    if f.NameLike != "" {
        conds = append(conds, "name LIKE ?")
        args = append(args, "%"+f.NameLike+"%")
    }
    if f.MinAge > 0 {
        conds = append(conds, "age >= ?")
        args = append(args, f.MinAge)
    }
    if f.MaxAge > 0 {
        conds = append(conds, "age <= ?")
        args = append(args, f.MaxAge)
    }

    where := ""
    if len(conds) > 0 {
        where = " WHERE " + strings.Join(conds, " AND ")
    }

    // Всего записей для метаданных пагинации
    var total int
    if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users"+where, args...).Scan(&total); err != nil {
        return nil, 0, err
    }

    // Сама выборка
    query := "SELECT id, name, email, age FROM users" + where + " ORDER BY id"
    queryArgs := append([]any{}, args...)
    if f.Limit > 0 {
        query += " LIMIT ? OFFSET ?"
        queryArgs = append(queryArgs, f.Limit, f.Offset)
    }

    rows, err := s.db.QueryContext(ctx, query, queryArgs...)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var u User
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Age); err != nil {
            return nil, 0, err
        }
        users = append(users, u)
    }
    return users, total, rows.Err()
}

func (s *SQLiteStorage) Get(ctx context.Context, id int64) (User, error) {
    var u User
    err := s.db.QueryRowContext(ctx,
        "SELECT id, name, email, age FROM users WHERE id = ?", id).
        Scan(&u.ID, &u.Name, &u.Email, &u.Age)
    if errors.Is(err, sql.ErrNoRows) {
        return User{}, ErrNotFound
    }
    return u, err
}

func (s *SQLiteStorage) Create(ctx context.Context, u User) (User, error) {
    res, err := s.db.ExecContext(ctx,
        "INSERT INTO users(name, email, age) VALUES(?, ?, ?)",
        u.Name, u.Email, u.Age)
    if err != nil {
        if isUniqueErr(err) {
            return User{}, ErrDuplicateEmail
        }
        return User{}, err
    }
    id, _ := res.LastInsertId()
    u.ID = id
    return u, nil
}

func (s *SQLiteStorage) Update(ctx context.Context, id int64, u User) (User, error) {
    res, err := s.db.ExecContext(ctx,
        "UPDATE users SET name=?, email=?, age=? WHERE id=?",
        u.Name, u.Email, u.Age, id)
    if err != nil {
        if isUniqueErr(err) {
            return User{}, ErrDuplicateEmail
        }
        return User{}, err
    }
    n, _ := res.RowsAffected()
    if n == 0 {
        return User{}, ErrNotFound
    }
    u.ID = id
    return u, nil
}

func (s *SQLiteStorage) Delete(ctx context.Context, id int64) error {
    res, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
    if err != nil {
        return err
    }
    n, _ := res.RowsAffected()
    if n == 0 {
        return ErrNotFound
    }
    return nil
}

func (s *SQLiteStorage) Close() error { return s.db.Close() }

func isUniqueErr(err error) bool {
    return strings.Contains(strings.ToLower(err.Error()), "unique")
}