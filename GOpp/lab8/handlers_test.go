package main

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strconv"
    "testing"

    "lab8/storage"
)

func newTestMux(t *testing.T) *http.ServeMux {
    t.Helper()
    h := NewUserHandler(storage.NewMemoryStorage())
    mux := http.NewServeMux()
    mux.Handle("GET /users", APIHandler(h.List))
    mux.Handle("POST /users", APIHandler(h.Create))
    mux.Handle("GET /users/{id}", APIHandler(h.Get))
    mux.Handle("PUT /users/{id}", APIHandler(h.Update))
    mux.Handle("DELETE /users/{id}", APIHandler(h.Delete))
    return mux
}

func doJSON(t *testing.T, mux *http.ServeMux, method, path string, body any) *httptest.ResponseRecorder {
    t.Helper()
    var buf bytes.Buffer
    if body != nil {
        json.NewEncoder(&buf).Encode(body)
    }
    req := httptest.NewRequest(method, path, &buf)
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()
    mux.ServeHTTP(rec, req)
    return rec
}

func TestCreateUser(t *testing.T) {
    cases := []struct {
        name      string
        body      any
        wantCode  int
    }{
        {"valid", CreateUserRequest{Name: "Alice", Email: "a@x.com", Age: 30}, 201},
        {"empty name", CreateUserRequest{Name: "", Email: "a@x.com", Age: 30}, 400},
        {"bad email", CreateUserRequest{Name: "A", Email: "not-email", Age: 30}, 400},
        {"negative age", CreateUserRequest{Name: "A", Email: "a@x.com", Age: -1}, 400},
        {"age too big", CreateUserRequest{Name: "A", Email: "a@x.com", Age: 200}, 400},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            mux := newTestMux(t)
            rec := doJSON(t, mux, "POST", "/users", tc.body)
            if rec.Code != tc.wantCode {
                t.Errorf("got %d, want %d, body: %s", rec.Code, tc.wantCode, rec.Body.String())
            }
        })
    }
}

func TestCreateDuplicateEmail(t *testing.T) {
    mux := newTestMux(t)
    body := CreateUserRequest{Name: "A", Email: "a@x.com", Age: 20}
    if r := doJSON(t, mux, "POST", "/users", body); r.Code != 201 {
        t.Fatalf("first create: %d", r.Code)
    }
    if r := doJSON(t, mux, "POST", "/users", body); r.Code != 409 {
        t.Errorf("duplicate: got %d, want 409", r.Code)
    }
}

func TestGetUser(t *testing.T) {
    mux := newTestMux(t)
    rec := doJSON(t, mux, "POST", "/users", CreateUserRequest{Name: "A", Email: "a@x.com", Age: 20})
    var u User
    json.Unmarshal(rec.Body.Bytes(), &u)

    // найден
    if r := doJSON(t, mux, "GET", "/users/"+itoa(u.ID), nil); r.Code != 200 {
        t.Errorf("got %d, want 200", r.Code)
    }
    // не найден
    if r := doJSON(t, mux, "GET", "/users/9999", nil); r.Code != 404 {
        t.Errorf("got %d, want 404", r.Code)
    }
    // невалидный id
    if r := doJSON(t, mux, "GET", "/users/abc", nil); r.Code != 400 {
        t.Errorf("got %d, want 400", r.Code)
    }
}

func TestUpdateUser(t *testing.T) {
    mux := newTestMux(t)
    rec := doJSON(t, mux, "POST", "/users", CreateUserRequest{Name: "A", Email: "a@x.com", Age: 20})
    var u User
    json.Unmarshal(rec.Body.Bytes(), &u)

    newName := "Updated"
    upd := UpdateUserRequest{Name: &newName}
    r := doJSON(t, mux, "PUT", "/users/"+itoa(u.ID), upd)
    if r.Code != 200 {
        t.Fatalf("got %d, want 200", r.Code)
    }
    var got User
    json.Unmarshal(r.Body.Bytes(), &got)
    if got.Name != "Updated" || got.Age != 20 {
        t.Errorf("update wrong: %+v", got)
    }

    // невалидный возраст
    badAge := 999
    if r := doJSON(t, mux, "PUT", "/users/"+itoa(u.ID), UpdateUserRequest{Age: &badAge}); r.Code != 400 {
        t.Errorf("got %d, want 400", r.Code)
    }
}

func TestDeleteUser(t *testing.T) {
    mux := newTestMux(t)
    rec := doJSON(t, mux, "POST", "/users", CreateUserRequest{Name: "A", Email: "a@x.com", Age: 20})
    var u User
    json.Unmarshal(rec.Body.Bytes(), &u)

    if r := doJSON(t, mux, "DELETE", "/users/"+itoa(u.ID), nil); r.Code != 204 {
        t.Errorf("got %d, want 204", r.Code)
    }
    if r := doJSON(t, mux, "DELETE", "/users/"+itoa(u.ID), nil); r.Code != 404 {
        t.Errorf("got %d, want 404", r.Code)
    }
}

func TestListPaginationFilter(t *testing.T) {
    mux := newTestMux(t)
    seed := []CreateUserRequest{
        {"Alice", "a@x.com", 25},
        {"Bob", "b@x.com", 30},
        {"Alex", "c@x.com", 35},
        {"Charlie", "d@x.com", 40},
    }
    for _, s := range seed {
        doJSON(t, mux, "POST", "/users", s)
    }

    // Фильтр по имени
    r := doJSON(t, mux, "GET", "/users?name=Al", nil)
    var resp struct {
        Data  []User `json:"data"`
        Total int    `json:"total"`
    }
    json.Unmarshal(r.Body.Bytes(), &resp)
    if resp.Total != 2 {
        t.Errorf("name=Al: total=%d, want 2", resp.Total)
    }

    // Фильтр по возрасту
    r = doJSON(t, mux, "GET", "/users?min_age=30&max_age=40", nil)
    json.Unmarshal(r.Body.Bytes(), &resp)
    if resp.Total != 3 {
        t.Errorf("age 30..40: total=%d, want 3", resp.Total)
    }

    // Пагинация
    r = doJSON(t, mux, "GET", "/users?page=2&limit=2", nil)
    json.Unmarshal(r.Body.Bytes(), &resp)
    if resp.Total != 4 || len(resp.Data) != 2 {
        t.Errorf("page=2 limit=2: total=%d, len=%d", resp.Total, len(resp.Data))
    }
}

// мини-замена strconv.FormatInt, чтобы не импортировать
func itoa(i int64) string { return strconv.FormatInt(i, 10) }