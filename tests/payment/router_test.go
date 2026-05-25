package tests

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gef3dx/it_courses/internal/course"
	"github.com/gef3dx/it_courses/internal/payment"
	"github.com/gef3dx/it_courses/internal/user"
	"github.com/gef3dx/it_courses/tests/testsupport"
)

func TestPaymentRouter_CreateAndCompleteGrantsCourseAccess(t *testing.T) {
	app, authSvc, userSvc, _, _, _ := testsupport.SetupTestApp(t)

	teacher := testsupport.SeedVerifiedUser(t, userSvc, user.RoleTeacher, "teacher2@example.com", "password123")
	teacherToken := testsupport.IssueAccessToken(t, authSvc, teacher)

	createCourseBody := testsupport.MustMarshal(t, course.CreateInput{
		Title:       "Algorithms",
		Description: "Algorithms course",
		Price:       2990,
		IsPublished: true,
	})
	req, _ := http.NewRequest(http.MethodPost, "/courses", bytes.NewReader(createCourseBody))
	req.Header.Set("Authorization", "Bearer "+teacherToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdCourse course.Model
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &createdCourse)

	student := testsupport.SeedVerifiedUser(t, userSvc, user.RoleStudent, "paying-student@example.com", "password123")
	studentToken := testsupport.IssueAccessToken(t, authSvc, student)

	createPaymentBody := testsupport.MustMarshal(t, payment.CreateInput{PaymentMethod: "card"})
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/courses/%d/payments", createdCourse.ID), bytes.NewReader(createPaymentBody))
	req.Header.Set("Authorization", "Bearer "+studentToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdPayment payment.Model
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &createdPayment)
	assert.Equal(t, createdCourse.Price, createdPayment.Amount)
	assert.Equal(t, payment.StatusPending, createdPayment.Status)

	req, _ = http.NewRequest(http.MethodGet, "/my/courses", nil)
	req.Header.Set("Authorization", "Bearer "+studentToken)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var myCourses []course.Model
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &myCourses)
	assert.Empty(t, myCourses)

	admin := testsupport.SeedVerifiedUser(t, userSvc, user.RoleAdmin, "admin2@example.com", "password123")
	adminToken := testsupport.IssueAccessToken(t, authSvc, admin)

	updateStatusBody := testsupport.MustMarshal(t, payment.UpdateStatusInput{
		Status:        payment.StatusCompleted,
		TransactionID: "txn-123",
	})
	req, _ = http.NewRequest(http.MethodPatch, fmt.Sprintf("/payments/%d/status", createdPayment.ID), bytes.NewReader(updateStatusBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated payment.Model
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &updated)
	assert.Equal(t, payment.StatusCompleted, updated.Status)
	require.NotNil(t, updated.PaidAt)

	req, _ = http.NewRequest(http.MethodGet, "/my/courses", nil)
	req.Header.Set("Authorization", "Bearer "+studentToken)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &myCourses)
	require.Len(t, myCourses, 1)
	assert.Equal(t, createdCourse.ID, myCourses[0].ID)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/payments/%d", createdPayment.ID), nil)
	req.Header.Set("Authorization", "Bearer "+studentToken)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
