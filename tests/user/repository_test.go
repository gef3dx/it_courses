package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gef3dx/it_courses/internal/user"
)

func TestRepository_CreateAndFindByEmail(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)

	created, err := repo.Create(context.Background(), user.CreateInput{
		Email:     "repo@example.com",
		Phone:     "9094445566",
		Name:      "Repo",
		FirstName: "Repo",
		LastName:  "User",
		Role:      user.RoleStudent,
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, user.RoleStudent, created.Role)

	found, err := repo.FindByEmail(context.Background(), "repo@example.com")
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestRepository_MarkEmailVerified(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)

	created, err := repo.Create(context.Background(), user.CreateInput{
		Email:                  "verify@example.com",
		Phone:                  "9094445566",
		Name:                   "Verify",
		FirstName:              "Verify",
		LastName:               "User",
		Role:                   user.RoleStudent,
		EmailVerificationToken: "verify-token",
	})
	require.NoError(t, err)

	verified, err := repo.MarkEmailVerified(context.Background(), created.ID)
	require.NoError(t, err)
	require.NotNil(t, verified.EmailVerifiedAt)
	assert.Empty(t, verified.EmailVerificationToken)
}
