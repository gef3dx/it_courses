package tests

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gef3dx/it_courses/internal/article"
	"github.com/gef3dx/it_courses/internal/user"
	"github.com/gef3dx/it_courses/tests/testsupport"
)

func TestArticleRouter_CreateAndGet(t *testing.T) {
	app, authSvc, userSvc, _, _, _ := testsupport.SetupTestApp(t)
	teacher := testsupport.SeedVerifiedUser(t, userSvc, user.RoleTeacher, "article-teacher@example.com", "password123")
	token := testsupport.IssueAccessToken(t, authSvc, teacher)

	body := testsupport.MustMarshal(t, article.CreateInput{
		Title:       "Linear Algebra",
		Content:     "rich text",
		IsPublished: true,
		Media: []article.MediaInput{{MediaType: "image", URL: "http://localhost:3000/uploads/a.png", SortOrder: 1}},
	})
	req, _ := http.NewRequest(http.MethodPost, "/articles", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created article.Model
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &created)
	require.Len(t, created.Media, 1)

	req, _ = http.NewRequest(http.MethodGet, "/articles/"+created.Slug, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
