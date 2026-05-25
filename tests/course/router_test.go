package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gef3dx/it_courses/internal/course"
	"github.com/gef3dx/it_courses/internal/user"
	"github.com/gef3dx/it_courses/tests/testsupport"
)

func TestCourseRouter_CreateAndGrantAccess(t *testing.T) {
	app, authSvc, userSvc, _, _, _ := testsupport.SetupTestApp(t)

	teacher := testsupport.SeedVerifiedUser(t, userSvc, user.RoleTeacher, "teacher@example.com", "password123")
	teacherToken := testsupport.IssueAccessToken(t, authSvc, teacher)

	createBody := testsupport.MustMarshal(t, course.CreateInput{
		Title:       "Go Basics",
		Description: "Intro course",
		Price:       1990,
		IsPublished: true,
	})

	req, _ := http.NewRequest(http.MethodPost, "/courses", bytes.NewReader(createBody))
	req.Header.Set("Authorization", "Bearer "+teacherToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created course.Model
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &created)
	assert.Equal(t, "go-basics", created.Slug)

	student := testsupport.SeedVerifiedUser(t, userSvc, user.RoleStudent, "student@example.com", "password123")
	studentToken := testsupport.IssueAccessToken(t, authSvc, student)

	req, _ = http.NewRequest(http.MethodGet, "/courses", nil)
	req.Header.Set("Authorization", "Bearer "+studentToken)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var courses []course.Model
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &courses)
	assert.Empty(t, courses)

	req, _ = http.NewRequest(http.MethodGet, "/courses/"+created.Slug, nil)
	req.Header.Set("Authorization", "Bearer "+studentToken)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	admin := testsupport.SeedVerifiedUser(t, userSvc, user.RoleAdmin, "admin@example.com", "password123")
	adminToken := testsupport.IssueAccessToken(t, authSvc, admin)

	grantBody := testsupport.MustMarshal(t, course.GrantAccessInput{UserID: student.ID})
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/courses/%d/access", created.ID), bytes.NewReader(grantBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req, _ = http.NewRequest(http.MethodGet, "/my/courses", nil)
	req.Header.Set("Authorization", "Bearer "+studentToken)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &courses)
	require.Len(t, courses, 1)
	assert.Equal(t, created.ID, courses[0].ID)

	req, _ = http.NewRequest(http.MethodGet, "/courses/"+created.Slug, nil)
	req.Header.Set("Authorization", "Bearer "+studentToken)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
