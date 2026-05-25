package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
	postgresdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gef3dx/it_courses/internal/auth"
	"github.com/gef3dx/it_courses/internal/config"
	"github.com/gef3dx/it_courses/internal/course"
	"github.com/gef3dx/it_courses/internal/database/postgres"
	"github.com/gef3dx/it_courses/internal/payment"
	"github.com/gef3dx/it_courses/internal/user"
)

type testMailer struct {
	messages []string
}

func (m *testMailer) Send(_ context.Context, email, subject, body string) error {
	m.messages = append(m.messages, email+"|"+subject+"|"+body)
	return nil
}

func getTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	dbName := getEnv("POSTGRES_DB", "it_courses")
	dbUser := getEnv("POSTGRES_USER", "it_user")
	password := getEnv("POSTGRES_PASSWORD", "it_password")
	sslmode := getEnv("POSTGRES_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, dbUser, password, dbName, sslmode)

	db, err := gorm.Open(postgresdriver.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("test database is unavailable: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql db: %v", err)
	}

	sqlDB.SetConnMaxLifetime(time.Minute)
	sqlDB.SetMaxOpenConns(2)
	sqlDB.SetMaxIdleConns(2)

	if err := sqlDB.Ping(); err != nil {
		t.Skipf("test database is unavailable: %v", err)
	}

	if err := postgres.ApplyMigrations(db, config.MigrationsConfig{AutoApply: true, Path: "../../migrations"}); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	return db
}

func getCleanDB(t *testing.T) *gorm.DB {
	t.Helper()

	db := getTestDB(t)

	t.Cleanup(func() {
		db.Exec("TRUNCATE TABLE password_reset_tokens RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	})

	db.Exec("TRUNCATE TABLE password_reset_tokens RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

	return db
}

func setupTestApp(t *testing.T) (*fiber.App, *auth.Service, *user.Service, *course.Service, *payment.Service, *testMailer) {
	t.Helper()

	db := getCleanDB(t)
	userRepo := user.NewRepository(db)
	authRepo := auth.NewRepository(db, userRepo)
	courseRepo := course.NewRepository(db)
	mailer := &testMailer{}
	authSvc := auth.NewService(authRepo, mailer, config.AuthConfig{
		JWTSecret:               "test-secret",
		AccessTokenTTLMinutes:   15,
		RefreshTokenTTLHours:    24,
		PasswordResetTTLMinutes: 60,
	})
	userSvc := user.NewService(userRepo)
	courseSvc := course.NewService(courseRepo)
	paymentSvc := payment.NewService(payment.NewRepository(db, courseRepo), courseRepo)

	app := fiber.New(fiber.Config{AppName: "test"})
	auth.RegisterRoutes(app, authSvc)
	user.RegisterRoutes(
		app,
		userSvc,
		authSvc.Required(),
		authSvc.Required(user.RoleAdmin),
		authSvc.OwnerOrAdmin(),
	)
	course.RegisterRoutes(app, courseSvc, authSvc)
	payment.RegisterRoutes(app, paymentSvc, authSvc)

	return app, authSvc, userSvc, courseSvc, paymentSvc, mailer
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

func seedUser(t *testing.T, svc *user.Service, overrides ...user.CreateInput) *user.Model {
	t.Helper()

	input := user.CreateInput{
		Email:     fmt.Sprintf("seed-%d@example.com", time.Now().UnixNano()),
		Phone:     "9094445566",
		Name:      "Seed",
		FirstName: "SeedFirst",
		LastName:  "SeedLast",
		Role:      user.RoleStudent,
	}
	if len(overrides) > 0 {
		input = overrides[0]
	}

	created, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("seed user failed: %v", err)
	}

	return created
}

func registerAndLogin(t *testing.T, app *fiber.App, mailer *testMailer, email string) auth.AuthResponse {
	t.Helper()

	registerBody := mustMarshal(t, auth.RegisterInput{
		Email:     email,
		Password:  "password123",
		Phone:     "9094445566",
		Name:      "User",
		FirstName: "First",
		LastName:  "Last",
	})

	req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(registerBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register returned status %d", resp.StatusCode)
	}

	_ = readBody(t, resp)
	if len(mailer.messages) == 0 {
		t.Fatalf("expected verification email to be sent")
	}

	token := mailer.messages[len(mailer.messages)-1]
	lastSeparator := bytes.LastIndexByte([]byte(token), '|')
	if lastSeparator == -1 {
		t.Fatalf("unexpected mail format")
	}

	verifyBody := mustMarshal(t, auth.VerifyEmailInput{Token: token[lastSeparator+1:]})
	req, _ = http.NewRequest(http.MethodPost, "/auth/verify-email", bytes.NewReader(verifyBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("verify request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("verify returned status %d", resp.StatusCode)
	}

	loginBody := mustMarshal(t, auth.LoginInput{
		Email:    email,
		Password: "password123",
	})
	req, _ = http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login returned status %d", resp.StatusCode)
	}

	var authResponse auth.AuthResponse
	mustUnmarshal(t, readBody(t, resp), &authResponse)
	return authResponse
}

func seedVerifiedUser(t *testing.T, userSvc *user.Service, role string, email string, password string) *user.Model {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password failed: %v", err)
	}

	now := time.Now().UTC()
	created, err := userSvc.Create(context.Background(), user.CreateInput{
		Email:           email,
		Phone:           "9094445566",
		Name:            "Seed",
		FirstName:       "Seed",
		LastName:        "User",
		PasswordHash:    string(hash),
		Role:            role,
		EmailVerifiedAt: &now,
	})
	if err != nil {
		t.Fatalf("seed verified user failed: %v", err)
	}

	return created
}

func issueAccessToken(t *testing.T, authSvc *auth.Service, model *user.Model) string {
	t.Helper()

	tokens, err := authSvc.IssueTokens(model)
	if err != nil {
		t.Fatalf("issue tokens failed: %v", err)
	}

	return tokens.AccessToken
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
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
