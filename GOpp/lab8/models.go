package main

// User — модель пользователя (то, что отдаём клиенту).
type User struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

// CreateUserRequest — тело POST /users.
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

// UpdateUserRequest — тело PUT /users/{id}.
// Указатели, чтобы отличить "поле не передано" от "передано нулевое значение".
type UpdateUserRequest struct {
    Name  *string `json:"name,omitempty"`
    Email *string `json:"email,omitempty"`
    Age   *int    `json:"age,omitempty"`
}