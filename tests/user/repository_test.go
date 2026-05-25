package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gef3dx/api_workinghub/internal/user"
)

func TestRepository_Create(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)

	created, err := repo.Create(context.Background(), user.CreateInput{
		Email:     "test@example.com",
		Phone:     "9094445566",
		Name:      "Test",
		FirstName: "TestFirst",
		LastName:  "TestLast",
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, int64(1), created.ID)
	assert.Equal(t, "test@example.com", created.Email)
	assert.Equal(t, "9094445566", created.Phone)
	assert.Equal(t, "Test", created.Name)
	assert.Equal(t, "TestFirst", created.FirstName)
	assert.Equal(t, "TestLast", created.LastName)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())
}

func TestRepository_Create_DuplicateEmail(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)

	input := user.CreateInput{
		Email: "dup@example.com", Phone: "9094445566",
		Name: "Test", FirstName: "Test", LastName: "Test",
	}

	_, err := repo.Create(context.Background(), input)
	require.NoError(t, err)

	_, err = repo.Create(context.Background(), input)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrUserConflict)
}

func TestRepository_Create_TrimsWhitespace(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)

	created, err := repo.Create(context.Background(), user.CreateInput{
		Email:     "  spaced@example.com  ",
		Phone:     "  9094445566  ",
		Name:      "  Spaced  ",
		FirstName: "  SpacedFirst  ",
		LastName:  "  SpacedLast  ",
	})

	require.NoError(t, err)
	assert.Equal(t, "spaced@example.com", created.Email)
	assert.Equal(t, "9094445566", created.Phone)
	assert.Equal(t, "Spaced", created.Name)
	assert.Equal(t, "SpacedFirst", created.FirstName)
	assert.Equal(t, "SpacedLast", created.LastName)
}

func TestRepository_List(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)
	ctx := context.Background()

	users, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, users)

	_, _ = repo.Create(ctx, user.CreateInput{
		Email: "first@example.com", Phone: "9094445566",
		Name: "A", FirstName: "A", LastName: "A",
	})
	_, _ = repo.Create(ctx, user.CreateInput{
		Email: "second@example.com", Phone: "9094445566",
		Name: "B", FirstName: "B", LastName: "B",
	})

	users, err = repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, int64(1), users[0].ID)
	assert.Equal(t, int64(2), users[1].ID)
}

func TestRepository_FindByID(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, user.CreateInput{
		Email: "findid@example.com", Phone: "9094445566",
		Name: "Find", FirstName: "Find", LastName: "Find",
	})
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.Email, found.Email)
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)

	_, err := repo.FindByID(context.Background(), 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestRepository_FindByEmail(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, user.CreateInput{
		Email: "findemail@example.com", Phone: "9094445566",
		Name: "Find", FirstName: "Find", LastName: "Find",
	})
	require.NoError(t, err)

	found, err := repo.FindByEmail(ctx, "findemail@example.com")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
}

func TestRepository_FindByEmail_NotFound(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)

	_, err := repo.FindByEmail(context.Background(), "nonexistent@example.com")
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestRepository_Update(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, user.CreateInput{
		Email: "update@example.com", Phone: "9094445566",
		Name: "OldName", FirstName: "OldFirst", LastName: "OldLast",
	})
	require.NoError(t, err)

	newName := "NewName"
	updated, err := repo.Update(ctx, created.ID, user.UpdateInput{Name: &newName})

	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "NewName", updated.Name)
	assert.Equal(t, "OldFirst", updated.FirstName)
	assert.Equal(t, "update@example.com", updated.Email)
	assert.True(t, updated.UpdatedAt.After(updated.CreatedAt))
}

func TestRepository_Update_AllFields(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, user.CreateInput{
		Email: "full@example.com", Phone: "9094445566",
		Name: "Old", FirstName: "Old", LastName: "Old",
	})
	require.NoError(t, err)

	email := "new@example.com"
	phone := "9112223344"
	name := "New"
	firstName := "NewFirst"
	lastName := "NewLast"

	updated, err := repo.Update(ctx, created.ID, user.UpdateInput{
		Email: &email, Phone: &phone, Name: &name,
		FirstName: &firstName, LastName: &lastName,
	})

	require.NoError(t, err)
	assert.Equal(t, "new@example.com", updated.Email)
	assert.Equal(t, "9112223344", updated.Phone)
	assert.Equal(t, "New", updated.Name)
	assert.Equal(t, "NewFirst", updated.FirstName)
	assert.Equal(t, "NewLast", updated.LastName)
}

func TestRepository_Update_NotFound(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)

	name := "Any"
	_, err := repo.Update(context.Background(), 999, user.UpdateInput{Name: &name})
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestRepository_Update_NoChanges(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, user.CreateInput{
		Email: "nochange@example.com", Phone: "9094445566",
		Name: "Same", FirstName: "Same", LastName: "Same",
	})
	require.NoError(t, err)

	updated, err := repo.Update(ctx, created.ID, user.UpdateInput{})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Same", updated.Name)
}

func TestRepository_Update_DuplicateEmail(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)
	ctx := context.Background()

	_, _ = repo.Create(ctx, user.CreateInput{
		Email: "existing@example.com", Phone: "9094445566",
		Name: "Existing", FirstName: "Existing", LastName: "Existing",
	})

	created, err := repo.Create(ctx, user.CreateInput{
		Email: "update@example.com", Phone: "9094445566",
		Name: "Update", FirstName: "Update", LastName: "Update",
	})
	require.NoError(t, err)

	dupEmail := "existing@example.com"
	_, err = repo.Update(ctx, created.ID, user.UpdateInput{Email: &dupEmail})
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrUserConflict)
}

func TestRepository_Delete(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, user.CreateInput{
		Email: "delete@example.com", Phone: "9094445566",
		Name: "Delete", FirstName: "Delete", LastName: "Delete",
	})
	require.NoError(t, err)

	err = repo.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, created.ID)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestRepository_Delete_NotFound(t *testing.T) {
	db := getCleanDB(t)
	repo := user.NewRepository(db)

	err := repo.Delete(context.Background(), 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, user.ErrUserNotFound)
}
