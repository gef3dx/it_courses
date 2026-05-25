package tests

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gef3dx/it_courses/internal/storage"
	"github.com/gef3dx/it_courses/internal/user"
	"github.com/gef3dx/it_courses/tests/testsupport"
)

func TestUploadRouter_Success(t *testing.T) {
	app, authSvc, userSvc, _, _, _ := testsupport.SetupTestApp(t)
	teacher := testsupport.SeedVerifiedUser(t, userSvc, user.RoleTeacher, "upload-teacher@example.com", "password123")
	token := testsupport.IssueAccessToken(t, authSvc, teacher)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "image.png")
	require.NoError(t, err)
	_, err = part.Write([]byte("png-content"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req, _ := http.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var uploaded storage.Object
	testsupport.MustUnmarshal(t, testsupport.ReadBody(t, resp), &uploaded)
	assert.Contains(t, uploaded.URL, "image.png")
}
