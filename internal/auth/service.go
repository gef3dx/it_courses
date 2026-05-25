package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/gef3dx/it_courses/internal/config"
	"github.com/gef3dx/it_courses/internal/mailer"
	"github.com/gef3dx/it_courses/internal/user"
)

type Service struct {
	repository *Repository
	sender     mailer.Sender
	validate   *validator.Validate
	config     config.AuthConfig
	now        func() time.Time
}

func NewService(repository *Repository, sender mailer.Sender, cfg config.AuthConfig) *Service {
	return &Service{
		repository: repository,
		sender:     sender,
		validate:   validator.New(),
		config:     cfg,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*user.Model, error) {
	input.Email = strings.TrimSpace(input.Email)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Name = strings.TrimSpace(input.Name)
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)

	if err := s.validate.Struct(input); err != nil {
		return nil, user.ValidationError{Message: firstValidationError(err)}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	verificationToken := uuid.NewString()

	created, err := s.repository.UserRepository().Create(ctx, user.CreateInput{
		Email:                  input.Email,
		Phone:                  input.Phone,
		Name:                   input.Name,
		FirstName:              input.FirstName,
		LastName:               input.LastName,
		PasswordHash:           string(passwordHash),
		Role:                   user.RoleStudent,
		EmailVerificationToken: verificationToken,
	})
	if err != nil {
		return nil, err
	}

	if err := s.sender.Send(ctx, created.Email, "Verify your email", verificationToken); err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) VerifyEmail(ctx context.Context, input VerifyEmailInput) (*user.Model, error) {
	if err := s.validate.Struct(input); err != nil {
		return nil, user.ValidationError{Message: firstValidationError(err)}
	}

	found, err := s.repository.UserRepository().FindByVerificationToken(ctx, input.Token)
	if err != nil {
		if errorsIsUserNotFound(err) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}

	return s.repository.UserRepository().MarkEmailVerified(ctx, found.ID)
}

func (s *Service) ResendVerification(ctx context.Context, input ResendVerificationInput) error {
	input.Email = strings.TrimSpace(input.Email)

	if err := s.validate.Struct(input); err != nil {
		return user.ValidationError{Message: firstValidationError(err)}
	}

	found, err := s.repository.UserRepository().FindByEmail(ctx, input.Email)
	if err != nil {
		return err
	}

	if found.EmailVerifiedAt != nil {
		return nil
	}

	token := uuid.NewString()
	if err := s.repository.UserRepository().SetVerificationToken(ctx, found.ID, token); err != nil {
		return err
	}

	return s.sender.Send(ctx, found.Email, "Verify your email", token)
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
	input.Email = strings.TrimSpace(input.Email)

	if err := s.validate.Struct(input); err != nil {
		return nil, user.ValidationError{Message: firstValidationError(err)}
	}

	found, err := s.repository.UserRepository().FindByEmail(ctx, input.Email)
	if err != nil {
		if errorsIsUserNotFound(err) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if found.EmailVerifiedAt == nil {
		return nil, ErrEmailNotVerified
	}

	if bcrypt.CompareHashAndPassword([]byte(found.PasswordHash), []byte(input.Password)) != nil {
		return nil, ErrInvalidCredentials
	}

	tokens, err := s.IssueTokens(found)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         *found,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, input RefreshInput) (*TokenPair, error) {
	if err := s.validate.Struct(input); err != nil {
		return nil, user.ValidationError{Message: firstValidationError(err)}
	}

	claims, err := s.ParseToken(input.RefreshToken, "refresh")
	if err != nil {
		return nil, err
	}

	found, err := s.repository.UserRepository().FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrUnauthorized
	}

	if found.TokenVersion != claims.TokenVersion {
		return nil, ErrUnauthorized
	}

	return s.IssueTokens(found)
}

func (s *Service) ForgotPassword(ctx context.Context, input ForgotPasswordInput) (*ResetTokenResponse, error) {
	input.Email = strings.TrimSpace(input.Email)

	if err := s.validate.Struct(input); err != nil {
		return nil, user.ValidationError{Message: firstValidationError(err)}
	}

	found, err := s.repository.UserRepository().FindByEmail(ctx, input.Email)
	if err != nil {
		if errorsIsUserNotFound(err) {
			return &ResetTokenResponse{}, nil
		}
		return nil, err
	}

	token := uuid.NewString()
	expiresAt := s.now().Add(s.config.PasswordResetTTL())

	resetToken, err := s.repository.CreatePasswordResetToken(ctx, found.ID, token, expiresAt)
	if err != nil {
		return nil, err
	}

	if err := s.sender.Send(ctx, found.Email, "Reset password", token); err != nil {
		return nil, err
	}

	return &ResetTokenResponse{
		Token:     resetToken.Token,
		ExpiresAt: resetToken.ExpiresAt,
	}, nil
}

func (s *Service) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	if err := s.validate.Struct(input); err != nil {
		return user.ValidationError{Message: firstValidationError(err)}
	}

	resetToken, err := s.repository.FindPasswordResetToken(ctx, input.Token)
	if err != nil {
		return err
	}

	if resetToken.ExpiresAt.Before(s.now()) {
		return ErrExpiredToken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.repository.UserRepository().UpdatePasswordHash(ctx, resetToken.UserID, string(passwordHash), true); err != nil {
		return err
	}

	return s.repository.DeletePasswordResetToken(ctx, input.Token)
}

func (s *Service) Required(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := strings.TrimSpace(c.Get("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "unauthorized"})
		}

		claims, err := s.ParseToken(strings.TrimPrefix(header, "Bearer "), "access")
		if err != nil {
			status := fiber.StatusUnauthorized
			if err == ErrForbidden {
				status = fiber.StatusForbidden
			}
			return c.Status(status).JSON(ErrorResponse{Error: err.Error()})
		}

		if len(roles) > 0 && !containsRole(roles, claims.Role) {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "forbidden"})
		}

		c.Locals("userID", claims.UserID)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

func (s *Service) OwnerOrAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, _ := c.Locals("userID").(int64)
		role, _ := c.Locals("role").(string)
		if role == user.RoleAdmin {
			return c.Next()
		}

		pathID, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid user id"})
		}

		if pathID != userID {
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "forbidden"})
		}

		return c.Next()
	}
}

func (s *Service) IssueTokens(model *user.Model) (*TokenPair, error) {
	accessToken, err := s.signToken(Claims{
		UserID:       model.ID,
		Role:         model.Role,
		TokenVersion: model.TokenVersion,
		TokenType:    "access",
		ExpiresAt:    s.now().Add(s.config.AccessTokenTTL()).Unix(),
	})
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.signToken(Claims{
		UserID:       model.ID,
		Role:         model.Role,
		TokenVersion: model.TokenVersion,
		TokenType:    "refresh",
		ExpiresAt:    s.now().Add(s.config.RefreshTokenTTL()).Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) ParseToken(token string, expectedType string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}

	expectedSig := s.sign(parts[0] + "." + parts[1])
	if !hmac.Equal([]byte(expectedSig), []byte(parts[2])) {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return nil, ErrInvalidToken
	}
	if claims.ExpiresAt < s.now().Unix() {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

func (s *Service) signToken(claims Claims) (string, error) {
	headerBytes, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}

	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	header := base64.RawURLEncoding.EncodeToString(headerBytes)
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := header + "." + payload

	return signingInput + "." + s.sign(signingInput), nil
}

func (s *Service) sign(value string) string {
	mac := hmac.New(sha256.New, []byte(s.config.JWTSecret))
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func firstValidationError(err error) string {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok || len(validationErrors) == 0 {
		return "invalid request data"
	}

	fieldErr := validationErrors[0]
	switch fieldErr.Field() {
	case "Email":
		if fieldErr.Tag() == "email" {
			return "email has invalid format"
		}
		return "email is required"
	case "Password":
		return "password length must be between 8 and 72 characters"
	case "NewPassword":
		return "new_password length must be between 8 and 72 characters"
	case "Token":
		return "token is required"
	case "RefreshToken":
		return "refresh_token is required"
	case "Phone":
		return "phone must contain only digits"
	case "Name":
		return "name length must be between 2 and 100 characters"
	case "FirstName":
		return "first_name length must be between 2 and 100 characters"
	case "LastName":
		return "last_name length must be between 2 and 100 characters"
	default:
		return "invalid request data"
	}
}

func containsRole(roles []string, role string) bool {
	for _, item := range roles {
		if item == role {
			return true
		}
	}
	return false
}

func errorsIsUserNotFound(err error) bool {
	return errors.Is(err, user.ErrUserNotFound)
}
