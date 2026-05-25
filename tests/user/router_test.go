package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gef3dx/api_workinghub/internal/user"
)

func TestRouter_ListUsers(t *testing.T) {
	app, svc := setupTestApp(t)

	seedUser(t, svc)
	seedUser(t, svc)

	req, _ := http.NewRequest(http.MethodGet, "/users", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var users []user.Model
	mustUnmarshal(t, readBody(t, resp), &users)
	assert.Len(t, users, 2)
}

func TestRouter_ListUsers_Empty(t *testing.T) {
	app, _ := setupTestApp(t)

	req, _ := http.NewRequest(http.MethodGet, "/users", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var users []user.Model
	mustUnmarshal(t, readBody(t, resp), &users)
	assert.Empty(t, users)
}

func TestRouter_CreateUser_Success(t *testing.T) {
	app, _ := setupTestApp(t)

	body := mustMarshal(t, user.CreateInput{
		Email: "new@example.com", Phone: "9094445566",
		Name: "New", FirstName: "NewFirst", LastName: "NewLast",
	})

	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created user.Model
	mustUnmarshal(t, readBody(t, resp), &created)
	assert.Equal(t, int64(1), created.ID)
	assert.Equal(t, "new@example.com", created.Email)
}

func TestRouter_CreateUser_ValidationError(t *testing.T) {
	app, _ := setupTestApp(t)

	body := mustMarshal(t, user.CreateInput{
		Email: "bad-email", Phone: "9094445566",
		Name: "Valid", FirstName: "Valid", LastName: "Valid",
	})

	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp user.ErrorResponse
	mustUnmarshal(t, readBody(t, resp), &errResp)
	assert.Contains(t, errResp.Error, "email has invalid format")
}

func TestRouter_CreateUser_Duplicate(t *testing.T) {
	app, _ := setupTestApp(t)

	body := mustMarshal(t, user.CreateInput{
		Email: "dup@example.com", Phone: "9094445566",
		Name: "Dup", FirstName: "Dup", LastName: "Dup",
	})

	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var errResp user.ErrorResponse
	mustUnmarshal(t, readBody(t, resp), &errResp)
	assert.Equal(t, "user with this email already exists", errResp.Error)
}

func TestRouter_CreateUser_InvalidBody(t *testing.T) {
	app, _ := setupTestApp(t)

	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRouter_GetUserByID_Success(t *testing.T) {
	app, svc := setupTestApp(t)
	created := seedUser(t, svc)

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", created.ID), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var found user.Model
	mustUnmarshal(t, readBody(t, resp), &found)
	assert.Equal(t, created.ID, found.ID)
}

func TestRouter_GetUserByID_InvalidID(t *testing.T) {
	app, _ := setupTestApp(t)

	req, _ := http.NewRequest(http.MethodGet, "/users/abc", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRouter_GetUserByID_NotFound(t *testing.T) {
	app, _ := setupTestApp(t)

	req, _ := http.NewRequest(http.MethodGet, "/users/999", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRouter_FindByEmail_Success(t *testing.T) {
	app, svc := setupTestApp(t)
	created := seedUser(t, svc)

	req, _ := http.NewRequest(http.MethodGet, "/users/by-email/"+created.Email, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var found user.Model
	mustUnmarshal(t, readBody(t, resp), &found)
	assert.Equal(t, created.ID, found.ID)
}

func TestRouter_FindByEmail_NotFound(t *testing.T) {
	app, _ := setupTestApp(t)

	req, _ := http.NewRequest(http.MethodGet, "/users/by-email/nobody@example.com", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRouter_UpdateUser_Success(t *testing.T) {
	app, svc := setupTestApp(t)
	created := seedUser(t, svc)

	body := mustMarshal(t, map[string]string{"name": "UpdatedName"})

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/users/%d", created.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated user.Model
	mustUnmarshal(t, readBody(t, resp), &updated)
	assert.Equal(t, "UpdatedName", updated.Name)
	assert.Equal(t, created.Email, updated.Email)
}

func TestRouter_UpdateUser_NotFound(t *testing.T) {
	app, _ := setupTestApp(t)

	body := mustMarshal(t, map[string]string{"name": "Any"})

	req, _ := http.NewRequest(http.MethodPut, "/users/999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRouter_UpdateUser_Conflict(t *testing.T) {
	app, svc := setupTestApp(t)

	seedUser(t, svc)

	_, err := svc.Create(context.Background(), user.CreateInput{
		Email: "other@example.com", Phone: "9094445566",
		Name: "Other", FirstName: "Other", LastName: "Other",
	})
	require.NoError(t, err)

	body := mustMarshal(t, map[string]string{"email": "other@example.com"})

	req, _ := http.NewRequest(http.MethodPut, "/users/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestRouter_DeleteUser_Success(t *testing.T) {
	app, svc := setupTestApp(t)
	created := seedUser(t, svc)

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", created.ID), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	_, err = svc.FindByEmail(context.Background(), created.Email)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestRouter_DeleteUser_NotFound(t *testing.T) {
	app, _ := setupTestApp(t)

	req, _ := http.NewRequest(http.MethodDelete, "/users/999", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestRouter_DeleteUser_InvalidID(t *testing.T) {
	app, _ := setupTestApp(t)

	req, _ := http.NewRequest(http.MethodDelete, "/users/abc", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRouter_UpdateUser_InvalidBody(t *testing.T) {
	app, svc := setupTestApp(t)
	seedUser(t, svc)

	body := mustMarshal(t, map[string]interface{}{
		"email": "not-an-email",
	})

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/users/%d", 1), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp user.ErrorResponse
	mustUnmarshal(t, readBody(t, resp), &errResp)
	assert.Equal(t, "email has invalid format", errResp.Error)
}

func TestRouter_DeleteUser_InvalidBody(t *testing.T) {
	app, _ := setupTestApp(t)

	req, _ := http.NewRequest(http.MethodDelete, "/users/abc", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRouter_CreateUser_MissingFields(t *testing.T) {
	app, _ := setupTestApp(t)

	tests := []struct {
		name string
		body map[string]string
	}{
		{"no_email", map[string]string{"phone": "9094445566", "name": "N", "first_name": "N", "last_name": "N"}},
		{"no_name", map[string]string{"email": "x@y.com", "phone": "9094445566", "first_name": "N", "last_name": "N"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.body)
			req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}
