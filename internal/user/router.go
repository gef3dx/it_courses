package user

import (
	"errors"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type userHandler struct {
	service *Service
}

func RegisterRoutes(app *fiber.App, service *Service, authRequired fiber.Handler, adminRequired fiber.Handler, ownerOrAdmin fiber.Handler) {
	handler := &userHandler{service: service}

	users := app.Group("/users", authRequired)
	users.Get("/", adminRequired, handler.list)
	users.Get("/:id", handler.getByID)
	users.Put("/:id", ownerOrAdmin, handler.update)
	users.Delete("/:id", ownerOrAdmin, handler.delete)
}

func (h *userHandler) list(c fiber.Ctx) error {
	users, err := h.service.List(c.Context())
	if err != nil {
		log.Printf("list users failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "failed to fetch users"})
	}

	return c.JSON(users)
}

func (h *userHandler) getByID(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid user id"})
	}

	model, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "user not found"})
		}

		log.Printf("get user by id failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "failed to fetch user"})
	}

	return c.JSON(model)
}

func (h *userHandler) update(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid user id"})
	}

	var input UpdateInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
	}

	model, err := h.service.Update(c.Context(), id, input)
	if err != nil {
		switch {
		case IsValidationError(err):
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: err.Error()})
		case errors.Is(err, ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "user not found"})
		case errors.Is(err, ErrUserConflict):
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{Error: "user with this email already exists"})
		default:
			log.Printf("update user failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "failed to update user"})
		}
	}

	return c.JSON(model)
}

func (h *userHandler) delete(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid user id"})
	}

	requestUserID, _ := c.Locals("userID").(int64)
	role, _ := c.Locals("role").(string)

	if requestUserID == id && role != RoleAdmin {
		var input DeleteInput
		if err := c.Bind().Body(&input); err != nil && c.Request().Header.ContentLength() > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request body"})
		}

		err = h.service.DeleteSelf(c.Context(), id, input.Password)
	} else {
		err = h.service.Delete(c.Context(), id)
	}

	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{Error: "user not found"})
		case errors.Is(err, ErrPasswordRequired):
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "password is required"})
		case errors.Is(err, ErrInvalidPassword):
			return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{Error: "invalid password"})
		case errors.Is(err, ErrLastAdmin):
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{Error: "cannot delete the last admin"})
		default:
			log.Printf("delete user failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: "failed to delete user"})
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}
