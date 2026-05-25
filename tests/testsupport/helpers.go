package testsupport

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
	"github.com/gef3dx/it_courses/internal/article"
	"github.com/gef3dx/it_courses/internal/config"
	"github.com/gef3dx/it_courses/internal/course"
	"github.com/gef3dx/it_courses/internal/database/postgres"
	"github.com/gef3dx/it_courses/internal/lesson"
	"github.com/gef3dx/it_courses/internal/page"
	"github.com/gef3dx/it_courses/internal/payment"
	"github.com/gef3dx/it_courses/internal/queue"
	"github.com/gef3dx/it_courses/internal/storage"
	testdomain "github.com/gef3dx/it_courses/internal/test"
	"github.com/gef3dx/it_courses/internal/upload"
	"github.com/gef3dx/it_courses/internal/user"
)

type TestMailer struct {
	Messages []string
}

func (m *TestMailer) Send(_ context.Context, email, subject, body string) error {
	m.Messages = append(m.Messages, email+"|"+subject+"|"+body)
	return nil
}

func SetupTestApp(t *testing.T) (*fiber.App, *auth.Service, *user.Service, *course.Service, *payment.Service, *TestMailer) {
	t.Helper()

	db := getCleanDB(t)
	userRepo := user.NewRepository(db)
	authRepo := auth.NewRepository(db, userRepo)
	courseRepo := course.NewRepository(db)
	mailer := &TestMailer{}
	authSvc := auth.NewService(authRepo, mailer, config.AuthConfig{
		JWTSecret:               "test-secret",
		AccessTokenTTLMinutes:   15,
		RefreshTokenTTLHours:    24,
		PasswordResetTTLMinutes: 60,
	})
	userSvc := user.NewService(userRepo)
	courseSvc := course.NewService(courseRepo)
	paymentSvc := payment.NewService(payment.NewRepository(db, courseRepo), courseRepo)
	testSvc := testdomain.NewService(testdomain.NewRepository(db))
	lessonSvc := lesson.NewService(lesson.NewRepository(db))
	pageSvc := page.NewService(page.NewRepository(db))
	articleSvc := article.NewService(article.NewRepository(db))
	storageSvc := storage.New("http://localhost:3000/uploads")
	publisher := queue.NoopPublisher{}

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
	testdomain.RegisterRoutes(app, testSvc, authSvc)
	lesson.RegisterRoutes(app, lessonSvc, authSvc)
	page.RegisterRoutes(app, pageSvc, authSvc)
	article.RegisterRoutes(app, articleSvc, authSvc)
	upload.RegisterRoutes(app, authSvc, storageSvc, publisher)

	return app, authSvc, userSvc, courseSvc, paymentSvc, mailer
}

func MustMarshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	return b
}

func MustUnmarshal(t *testing.T, data []byte, v interface{}) {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
}

func ReadBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(resp.Body)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}
	_ = resp.Body.Close()
	return buf.Bytes()
}

func SeedVerifiedUser(t *testing.T, userSvc *user.Service, role string, email string, password string) *user.Model {
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

func IssueAccessToken(t *testing.T, authSvc *auth.Service, model *user.Model) string {
	t.Helper()

	tokens, err := authSvc.IssueTokens(model)
	if err != nil {
		t.Fatalf("issue tokens failed: %v", err)
	}

	return tokens.AccessToken
}

func GetEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	host := GetEnv("POSTGRES_HOST", "localhost")
	port := GetEnv("POSTGRES_PORT", "5432")
	dbName := GetEnv("POSTGRES_DB", "it_courses")
	dbUser := GetEnv("POSTGRES_USER", "it_user")
	password := GetEnv("POSTGRES_PASSWORD", "it_password")
	sslmode := GetEnv("POSTGRES_SSLMODE", "disable")

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
		db.Exec("TRUNCATE TABLE article_test_links RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE article_media RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE articles RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE pages RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE test_answers RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE test_results RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE answer_options RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE questions RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE lesson_test_links RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE lesson_media RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE lessons RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE tests RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE payments RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE course_accesses RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE courses RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE password_reset_tokens RESTART IDENTITY CASCADE")
		db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	})

	db.Exec("TRUNCATE TABLE article_test_links RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE article_media RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE articles RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE pages RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE test_answers RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE test_results RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE answer_options RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE questions RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE lesson_test_links RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE lesson_media RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE lessons RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE tests RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE payments RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE course_accesses RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE courses RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE password_reset_tokens RESTART IDENTITY CASCADE")
	db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE")

	return db
}
