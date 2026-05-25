package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testdomain "github.com/gef3dx/it_courses/internal/test"
	"github.com/gef3dx/it_courses/internal/user"
	"github.com/gef3dx/it_courses/tests/testsupport"
)

func TestTestRouter_CreateQuestionAndSubmit(t *testing.T) {
	app, authSvc, userSvc, _, _, _ := testsupport.SetupTestApp(t)
	teacher := testsupport.SeedVerifiedUser(t, userSvc, user.RoleTeacher, "test-teacher@example.com", "password123")
	teacherToken := testsupport.IssueAccessToken(t, authSvc, teacher)

	createBody := testsupport.MustMarshal(t, testdomain.CreateInput{Title: "Go Quiz", Description: "quiz"})
	req, _ := http.NewRequest(http.MethodPost, "/tests", bytes.NewReader(createBody))
	req.Header.Set("Authorization", "Bearer "+teacherToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created testdomain.Model
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &created)

	questionBody := testsupport.MustMarshal(t, testdomain.QuestionInput{
		Text: "2+2", Solution: "4", SortOrder: 1,
		Options: []testdomain.AnswerOptionInput{
			{Text: "3"}, {Text: "4", IsCorrect: true},
		},
	})
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/tests/%d/questions", created.ID), bytes.NewReader(questionBody))
	req.Header.Set("Authorization", "Bearer "+teacherToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var question testdomain.Question
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &question)

	student := testsupport.SeedVerifiedUser(t, userSvc, user.RoleStudent, "test-student@example.com", "password123")
	studentToken := testsupport.IssueAccessToken(t, authSvc, student)

	submitBody := testsupport.MustMarshal(t, testdomain.SubmitInput{
		Answers: []testdomain.SubmitAnswerInput{{QuestionID: question.ID, SelectedOptionID: question.Options[1].ID}},
	})
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/tests/%d/submit", created.ID), bytes.NewReader(submitBody))
	req.Header.Set("Authorization", "Bearer "+studentToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result testdomain.Result
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &result)
	assert.Equal(t, 1, result.Score)
	assert.Equal(t, 100.0, result.Grade)
}
