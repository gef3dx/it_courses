package auth

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"github.com/gef3dx/it_courses/internal/user"
)

type handler struct {
	service *Service
}

func RegisterRoutes(app *fiber.App, service *Service) {
	h := &handler{service: service}
	authGroup := app.Group("/auth")

	authGroup.Post("/register", h.register)
	authGroup.Post("/verify-email", h.verifyEmail)
	authGroup.Post("/resend-verification", h.resendVerification)
	authGroup.Post("/login", h.login)
	authGroup.Post("/refresh", h.refresh)
	authGroup.Post("/forgot-password", h.forgotPassword)
	authGroup.Post("/reset-password", h.resetPassword)
}

func (h *handler) register(c fiber.Ctx) error {
	var input RegisterInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	created, err := h.service.Register(c.Context(), input)
	if err != nil {
		return authError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(created)
}

func (h *handler) verifyEmail(c fiber.Ctx) error {
	var input VerifyEmailInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	verified, err := h.service.VerifyEmail(c.Context(), input)
	if err != nil {
		return authError(c, err)
	}

	return c.JSON(verified)
}

func (h *handler) resendVerification(c fiber.Ctx) error {
	var input ResendVerificationInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	if err := h.service.ResendVerification(c.Context(), input); err != nil {
		return authError(c, err)
	}

	return c.JSON(MessageResponse{Message: "verification email sent"})
}

func (h *handler) login(c fiber.Ctx) error {
	var input LoginInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	response, err := h.service.Login(c.Context(), input)
	if err != nil {
		return authError(c, err)
	}

	return c.JSON(response)
}

func (h *handler) refresh(c fiber.Ctx) error {
	var input RefreshInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	tokens, err := h.service.Refresh(c.Context(), input)
	if err != nil {
		return authError(c, err)
	}

	return c.JSON(tokens)
}

func (h *handler) forgotPassword(c fiber.Ctx) error {
	var input ForgotPasswordInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	response, err := h.service.ForgotPassword(c.Context(), input)
	if err != nil {
		return authError(c, err)
	}

	return c.JSON(response)
}

func (h *handler) resetPassword(c fiber.Ctx) error {
	var input ResetPasswordInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	if err := h.service.ResetPassword(c.Context(), input); err != nil {
		return authError(c, err)
	}

	return c.JSON(MessageResponse{Message: "password reset successful"})
}

func authError(c fiber.Ctx, err error) error {
	switch {
	case user.IsValidationError(err):
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
	case errors.Is(err, user.ErrUserConflict):
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{Error: "user with this email already exists"})
	case errors.Is(err, user.ErrUserNotFound), errors.Is(err, ErrInvalidToken), errors.Is(err, ErrInvalidCredentials):
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	case errors.Is(err, ErrEmailNotVerified):
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: err.Error()})
	case errors.Is(err, ErrExpiredToken):
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "internal server error"})
	}
}
