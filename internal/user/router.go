package user

import (
	"errors"
	"log"
	"net/url"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// userHandler хранит зависимости HTTP-слоя user-модуля.
type userHandler struct {
	service *Service
}

// RegisterRoutes регистрирует HTTP-обработчики user-модуля в основном приложении.
func RegisterRoutes(app *fiber.App, service *Service) {
	handler := &userHandler{service: service}

	app.Get("/users", handler.list)
	app.Post("/users", handler.create)
	app.Get("/users/by-email/:email", handler.findByEmail)
	app.Get("/users/:id", handler.getByID)
	app.Put("/users/:id", handler.update)
	app.Delete("/users/:id", handler.delete)
}

// list возвращает список всех пользователей из базы данных.
//
// @Summary Получить список пользователей
// @Description Возвращает всех пользователей из базы данных.
// @Tags users
// @Produce json
// @Success 200 {array} Model
// @Failure 500 {object} ErrorResponse
// @Router /users [get]
func (h *userHandler) list(c fiber.Ctx) error {
	users, err := h.service.List(c.Context())
	if err != nil {
		log.Printf("list users failed: %v", err)

		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "failed to fetch users",
		})
	}

	return c.JSON(users)
}

// getByID возвращает пользователя по ID.
//
// @Summary Получить пользователя по ID
// @Description Возвращает пользователя по его идентификатору.
// @Tags users
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} Model
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [get]
func (h *userHandler) getByID(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid user id",
		})
	}

	user, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error: "user not found",
			})
		}
		log.Printf("get user by id failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "failed to fetch user",
		})
	}

	return c.JSON(user)
}

// findByEmail возвращает пользователя по email.
//
// @Summary Получить пользователя по email
// @Description Возвращает пользователя по его email.
// @Tags users
// @Produce json
// @Param email path string true "Email пользователя"
// @Success 200 {object} Model
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/by-email/{email} [get]
func (h *userHandler) findByEmail(c fiber.Ctx) error {
	email, err := url.PathUnescape(c.Params("email"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid email encoding",
		})
	}
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "email is required",
		})
	}

	user, err := h.service.FindByEmail(c.Context(), email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error: "user not found",
			})
		}
		log.Printf("find user by email failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "failed to fetch user",
		})
	}

	return c.JSON(user)
}

// create создаёт нового пользователя по JSON-телу запроса.
//
// @Summary Создать пользователя
// @Description Создаёт нового пользователя в базе данных.
// @Tags users
// @Accept json
// @Produce json
// @Param payload body CreateInput true "Данные пользователя"
// @Success 201 {object} Model
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [post]
func (h *userHandler) create(c fiber.Ctx) error {
	var input CreateInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid request body",
		})
	}

	user, err := h.service.Create(c.Context(), input)
	if err != nil {
		switch {
		case IsValidationError(err):
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error: err.Error(),
			})
		case errors.Is(err, ErrUserConflict):
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Error: "user with this email already exists",
			})
		default:
			log.Printf("create user failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Error: "failed to create user",
			})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

// update обновляет пользователя по ID.
//
// @Summary Обновить пользователя
// @Description Обновляет поля пользователя по ID.
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Param payload body UpdateInput true "Данные для обновления"
// @Success 200 {object} Model
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [put]
func (h *userHandler) update(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid user id",
		})
	}

	var input UpdateInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid request body",
		})
	}

	user, err := h.service.Update(c.Context(), id, input)
	if err != nil {
		switch {
		case IsValidationError(err):
			return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
				Error: err.Error(),
			})
		case errors.Is(err, ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error: "user not found",
			})
		case errors.Is(err, ErrUserConflict):
			return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
				Error: "user with this email already exists",
			})
		default:
			log.Printf("update user failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Error: "failed to update user",
			})
		}
	}

	return c.JSON(user)
}

// delete удаляет пользователя по ID.
//
// @Summary Удалить пользователя
// @Description Удаляет пользователя по ID.
// @Tags users
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [delete]
func (h *userHandler) delete(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "invalid user id",
		})
	}

	if err := h.service.Delete(c.Context(), id); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
				Error: "user not found",
			})
		default:
			log.Printf("delete user failed: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
				Error: "failed to delete user",
			})
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}
