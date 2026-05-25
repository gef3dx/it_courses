package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gef3dx/it_courses/internal/auth"
	"github.com/gef3dx/it_courses/internal/user"
)

func TestRouter_RegisterVerifyLoginRefresh(t *testing.T) {
	app, _, _, _, _, mailer := setupTestApp(t)

	registerBody := mustMarshal(t, auth.RegisterInput{
		Email:     "router@example.com",
		Password:  "password123",
		Phone:     "9094445566",
		Name:      "Router",
		FirstName: "Route",
		LastName:  "Tester",
	})

	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Len(t, mailer.messages, 1)

	_ = readBody(t, resp)
	message := mailer.messages[0]
	lastSeparator := bytes.LastIndexByte([]byte(message), '|')
	require.NotEqual(t, -1, lastSeparator)

	verifyBody := mustMarshal(t, auth.VerifyEmailInput{Token: message[lastSeparator+1:]})
	req, _ = http.NewRequest(http.MethodPost, "/auth/verify-email", bytes.NewReader(verifyBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	loginBody := mustMarshal(t, auth.LoginInput{Email: "router@example.com", Password: "password123"})
	req, _ = http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var authResponse auth.AuthResponse
	mustUnmarshal(t, readBody(t, resp), &authResponse)
	assert.NotEmpty(t, authResponse.AccessToken)
	assert.NotEmpty(t, authResponse.RefreshToken)

	refreshBody := mustMarshal(t, auth.RefreshInput{RefreshToken: authResponse.RefreshToken})
	req, _ = http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(refreshBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRouter_ProtectedUsersEndpoints(t *testing.T) {
	app, _, _, _, _, mailer := setupTestApp(t)
	authResponse := registerAndLogin(t, app, mailer, "protected@example.com")

	req, _ := http.NewRequest(http.MethodGet, "/users", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/users/%d", authResponse.User.ID), nil)
	req.Header.Set("Authorization", "Bearer "+authResponse.AccessToken)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	updateBody := mustMarshal(t, map[string]string{"name": "Updated"})
	req, _ = http.NewRequest(http.MethodPut, fmt.Sprintf("/users/%d", authResponse.User.ID), bytes.NewReader(updateBody))
	req.Header.Set("Authorization", "Bearer "+authResponse.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRouter_DeleteSelfRequiresPassword(t *testing.T) {
	app, _, _, _, _, mailer := setupTestApp(t)
	authResponse := registerAndLogin(t, app, mailer, "delete-self@example.com")

	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", authResponse.User.ID), nil)
	req.Header.Set("Authorization", "Bearer "+authResponse.AccessToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	deleteBody := mustMarshal(t, user.DeleteInput{Password: "password123"})
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%d", authResponse.User.ID), bytes.NewReader(deleteBody))
	req.Header.Set("Authorization", "Bearer "+authResponse.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
