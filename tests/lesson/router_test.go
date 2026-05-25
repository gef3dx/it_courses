package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gef3dx/it_courses/internal/course"
	"github.com/gef3dx/it_courses/internal/lesson"
	"github.com/gef3dx/it_courses/internal/user"
	"github.com/gef3dx/it_courses/tests/testsupport"
)

func TestLessonRouter_CreateQuestionAndSubmit(t *testing.T) {
	app, authSvc, userSvc, _, _, _ := testsupport.SetupTestApp(t)
	teacher := testsupport.SeedVerifiedUser(t, userSvc, user.RoleTeacher, "lesson-teacher@example.com", "password123")
	teacherToken := testsupport.IssueAccessToken(t, authSvc, teacher)

	courseBody := testsupport.MustMarshal(t, course.CreateInput{Title: "Physics", Price: 0, IsPublished: true})
	req, _ := http.NewRequest(http.MethodPost, "/courses", bytes.NewReader(courseBody))
	req.Header.Set("Authorization", "Bearer "+teacherToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdCourse course.Model
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &createdCourse)

	lessonBody := testsupport.MustMarshal(t, lesson.CreateInput{Title: "Lesson 1", Content: "content", SortOrder: 1, IsPublished: true})
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/courses/%d/lessons", createdCourse.ID), bytes.NewReader(lessonBody))
	req.Header.Set("Authorization", "Bearer "+teacherToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdLesson lesson.Model
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &createdLesson)

	questionBody := testsupport.MustMarshal(t, lesson.QuestionInput{
		Text: "Sun is a?", Solution: "star", SortOrder: 1,
		Options: []lesson.AnswerOptionInput{
			{Text: "planet"}, {Text: "star", IsCorrect: true},
		},
	})
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/lessons/%d/questions", createdLesson.ID), bytes.NewReader(questionBody))
	req.Header.Set("Authorization", "Bearer "+teacherToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var question lesson.Question
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &question)

	student := testsupport.SeedVerifiedUser(t, userSvc, user.RoleStudent, "lesson-student@example.com", "password123")
	studentToken := testsupport.IssueAccessToken(t, authSvc, student)

	submitBody := testsupport.MustMarshal(t, lesson.SubmitInput{
		Answers: []lesson.SubmitAnswerInput{{QuestionID: question.ID, SelectedOptionID: question.Options[1].ID}},
	})
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/lessons/%d/submit", createdLesson.ID), bytes.NewReader(submitBody))
	req.Header.Set("Authorization", "Bearer "+studentToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result lesson.SubmitResponse
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &result)
	assert.Equal(t, 1, result.Score)
	assert.Equal(t, 100.0, result.Grade)
}
