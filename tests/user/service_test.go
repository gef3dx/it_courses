package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/gef3dx/it_courses/internal/user"
)

func TestService_Create_ValidationErrors(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))

	for _, tc := range validationCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tc.input)
			require.Error(t, err)
			assert.True(t, user.IsValidationError(err))
			assert.Equal(t, tc.want, err.Error())
		})
	}
}

func TestService_DeleteSelf_RequiresPassword(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	created, err := svc.Create(context.Background(), user.CreateInput{
		Email:        "self@example.com",
		Phone:        "9094445566",
		Name:         "Self",
		FirstName:    "Self",
		LastName:     "User",
		PasswordHash: string(passwordHash),
		Role:         user.RoleStudent,
	})
	require.NoError(t, err)

	err = svc.DeleteSelf(context.Background(), created.ID, "")
	assert.ErrorIs(t, err, user.ErrPasswordRequired)

	err = svc.DeleteSelf(context.Background(), created.ID, "wrong-password")
	assert.ErrorIs(t, err, user.ErrInvalidPassword)

	require.NoError(t, svc.DeleteSelf(context.Background(), created.ID, "password123"))
}
