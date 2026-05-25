package tests

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gef3dx/it_courses/internal/page"
	"github.com/gef3dx/it_courses/internal/user"
	"github.com/gef3dx/it_courses/tests/testsupport"
)

func TestPageRouter_CreateAndGet(t *testing.T) {
	app, authSvc, userSvc, _, _, _ := testsupport.SetupTestApp(t)
	teacher := testsupport.SeedVerifiedUser(t, userSvc, user.RoleTeacher, "page-teacher@example.com", "password123")
	token := testsupport.IssueAccessToken(t, authSvc, teacher)

	body := testsupport.MustMarshal(t, page.CreateInput{Title: "About School", Content: "content", IsPublished: true})
	req, _ := http.NewRequest(http.MethodPost, "/pages", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created page.Model
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &created)
	assert.Equal(t, "about-school", created.Slug)

	req, _ = http.NewRequest(http.MethodGet, "/pages/"+created.Slug, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
