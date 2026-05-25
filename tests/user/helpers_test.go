package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/gofiber/fiber/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gef3dx/it_courses/internal/user"
)

func getTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	dbName := getEnv("POSTGRES_DB", "wh")
	dbUser := getEnv("POSTGRES_USER", "wh_user")
	password := getEnv("POSTGRES_PASSWORD", "wh_password")
	sslmode := getEnv("POSTGRES_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, dbUser, password, dbName, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	_ = db.AutoMigrate(&user.Model{})

	return db
}

func getCleanDB(t *testing.T) *gorm.DB {
	t.Helper()

	db := getTestDB(t)

	t.Cleanup(func() {
		db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	})

	db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

	return db
}

func setupTestApp(t *testing.T) (*fiber.App, *user.Service) {
	t.Helper()

	db := getCleanDB(t)
	repo := user.NewRepository(db)
	svc := user.NewService(repo)

	app := fiber.New(fiber.Config{AppName: "test"})
	user.RegisterRoutes(app, svc)

	return app, svc
}

func mustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	return b
}

func mustUnmarshal(t *testing.T, data []byte, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
}

func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}
	_ = resp.Body.Close()
	return buf.Bytes()
}

var seedCounter int

func seedUser(t *testing.T, svc *user.Service) *user.Model {
	t.Helper()

	seedCounter++
	email := fmt.Sprintf("seed%d@example.com", seedCounter)

	created, err := svc.Create(context.Background(), user.CreateInput{
		Email:     email,
		Phone:     "9094445566",
		Name:      "Seed",
		FirstName: "SeedFirst",
		LastName:  "SeedLast",
	})
	if err != nil {
		t.Fatalf("seed user failed: %v", err)
	}
	return created
}

var validationCases = []struct {
	name  string
	input user.CreateInput
	want  string
}{
	{
		name: "missing email",
		input: user.CreateInput{
			Phone: "9094445566", Name: "Valid", FirstName: "Valid", LastName: "Valid",
		},
		want: "email is required",
	},
	{
		name: "invalid email",
		input: user.CreateInput{
			Email: "bad-email", Phone: "9094445566",
			Name: "Valid", FirstName: "Valid", LastName: "Valid",
		},
		want: "email has invalid format",
	},
	{
		name: "missing phone",
		input: user.CreateInput{
			Email: "test@example.com", Name: "Valid", FirstName: "Valid", LastName: "Valid",
		},
		want: "phone is required",
	},
	{
		name: "non-numeric phone",
		input: user.CreateInput{
			Email: "test@example.com", Phone: "abc",
			Name: "Valid", FirstName: "Valid", LastName: "Valid",
		},
		want: "phone must contain only digits",
	},
	{
		name: "short phone",
		input: user.CreateInput{
			Email: "test@example.com", Phone: "123",
			Name: "Valid", FirstName: "Valid", LastName: "Valid",
		},
		want: "phone length must be between 10 and 20 digits",
	},
	{
		name: "short name",
		input: user.CreateInput{
			Email: "test@example.com", Phone: "9094445566",
			Name: "A", FirstName: "Valid", LastName: "Valid",
		},
		want: "name length must be between 2 and 100 characters",
	},
	{
		name: "missing name",
		input: user.CreateInput{
			Email: "test@example.com", Phone: "9094445566",
			FirstName: "Valid", LastName: "Valid",
		},
		want: "name is required",
	},
	{
		name: "missing first_name",
		input: user.CreateInput{
			Email: "test@example.com", Phone: "9094445566",
			Name: "Valid", LastName: "Valid",
		},
		want: "first_name is required",
	},
	{
		name: "missing last_name",
		input: user.CreateInput{
			Email: "test@example.com", Phone: "9094445566",
			Name: "Valid", FirstName: "Valid",
		},
		want: "last_name is required",
	},
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
