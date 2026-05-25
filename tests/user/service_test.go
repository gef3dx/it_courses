package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gef3dx/it_courses/internal/user"
)

func TestService_Create_Success(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))

	created, err := svc.Create(context.Background(), user.CreateInput{
		Email: "svc@example.com", Phone: "9094445566",
		Name: "Svc", FirstName: "SvcFirst", LastName: "SvcLast",
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, int64(1), created.ID)
	assert.Equal(t, "svc@example.com", created.Email)
}

func TestService_Create_NormalizesInput(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))

	created, err := svc.Create(context.Background(), user.CreateInput{
		Email:     "  NORMALIZE@example.com  ",
		Phone:     "  9094445566  ",
		Name:      "  Normalize  ",
		FirstName: "  NormalizeFirst  ",
		LastName:  "  NormalizeLast  ",
	})

	require.NoError(t, err)
	assert.Equal(t, "NORMALIZE@example.com", created.Email)
	assert.Equal(t, "9094445566", created.Phone)
	assert.Equal(t, "Normalize", created.Name)
	assert.Equal(t, "NormalizeFirst", created.FirstName)
	assert.Equal(t, "NormalizeLast", created.LastName)
}

func TestService_Create_ValidationErrors(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))
	ctx := context.Background()

	for _, tc := range validationCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(ctx, tc.input)
			require.Error(t, err)
			assert.True(t, user.IsValidationError(err))
			assert.Equal(t, tc.want, err.Error())
		})
	}
}

func TestService_Create_DuplicateEmail(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))
	ctx := context.Background()

	input := user.CreateInput{
		Email: "dup-svc@example.com", Phone: "9094445566",
		Name: "Dup", FirstName: "Dup", LastName: "Dup",
	}

	_, err := svc.Create(ctx, input)
	require.NoError(t, err)

	_, err = svc.Create(ctx, input)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrUserConflict)
}

func TestService_Update_Success(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))
	ctx := context.Background()

	created, err := svc.Create(ctx, user.CreateInput{
		Email: "svc-upd@example.com", Phone: "9094445566",
		Name: "Old", FirstName: "Old", LastName: "Old",
	})
	require.NoError(t, err)

	newName := "Updated"
	updated, err := svc.Update(ctx, created.ID, user.UpdateInput{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
	assert.Equal(t, "Old", updated.FirstName)
}

func TestService_Update_NormalizesInput(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))
	ctx := context.Background()

	created, err := svc.Create(ctx, user.CreateInput{
		Email: "svc-norm@example.com", Phone: "9094445566",
		Name: "Norm", FirstName: "Norm", LastName: "Norm",
	})
	require.NoError(t, err)

	newName := "  SpacedName  "
	updated, err := svc.Update(ctx, created.ID, user.UpdateInput{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, "SpacedName", updated.Name)
}

func TestService_Update_NotFound(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))

	name := "Any"
	_, err := svc.Update(context.Background(), 999, user.UpdateInput{Name: &name})
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestService_Delete_Success(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))
	ctx := context.Background()

	created, err := svc.Create(ctx, user.CreateInput{
		Email: "svc-del@example.com", Phone: "9094445566",
		Name: "Del", FirstName: "Del", LastName: "Del",
	})
	require.NoError(t, err)

	err = svc.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = svc.FindByEmail(ctx, "svc-del@example.com")
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestService_Delete_NotFound(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))

	err := svc.Delete(context.Background(), 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestService_List(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))
	ctx := context.Background()

	users, err := svc.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, users)

	_, err = svc.Create(ctx, user.CreateInput{
		Email: "list1@example.com", Phone: "9094445566",
		Name: "UserA", FirstName: "FirstA", LastName: "LastA",
	})
	require.NoError(t, err)

	_, err = svc.Create(ctx, user.CreateInput{
		Email: "list2@example.com", Phone: "9094445566",
		Name: "UserB", FirstName: "FirstB", LastName: "LastB",
	})
	require.NoError(t, err)

	users, err = svc.List(ctx)
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestService_GetByID(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))
	ctx := context.Background()

	created, err := svc.Create(ctx, user.CreateInput{
		Email: "getbyid@example.com", Phone: "9094445566",
		Name: "Get", FirstName: "Get", LastName: "Get",
	})
	require.NoError(t, err)

	found, err := svc.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestService_GetByID_NotFound(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))

	_, err := svc.GetByID(context.Background(), 999)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestService_FindByEmail(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))
	ctx := context.Background()

	created, err := svc.Create(ctx, user.CreateInput{
		Email: "svc-find@example.com", Phone: "9094445566",
		Name: "Find", FirstName: "Find", LastName: "Find",
	})
	require.NoError(t, err)

	found, err := svc.FindByEmail(ctx, "svc-find@example.com")
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
}

func TestService_FindByEmail_NotFound(t *testing.T) {
	db := getCleanDB(t)
	svc := user.NewService(user.NewRepository(db))

	_, err := svc.FindByEmail(context.Background(), "nonexistent@example.com")
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}
